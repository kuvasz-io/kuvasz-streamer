package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
)

func pluginArguments(pgVersion int, slotName string) (int, []string) {
	protocolVersion := 2
	arg := make([]string, 0, 10)
	switch pgVersion {
	case 12, 13:
		arg = append(arg, "proto_version '1'")
		protocolVersion = 1
	case 14:
		arg = append(arg, "proto_version '2'")
	default:
		arg = append(arg, "proto_version '2'")
	}
	arg = append(arg, fmt.Sprintf("publication_names '%s'", slotName))
	if pgVersion > 13 {
		arg = append(arg, "binary 'false'")
		arg = append(arg, "messages 'true'")
		arg = append(arg, "streaming 'true'")
	}
	return protocolVersion, arg
}

func createReplicationSlot(
	log *slog.Logger,
	conn *pgx.Conn,
	slotName string,
	xlogpos pglogrepl.LSN) (bool, pglogrepl.LSN, error) {
	ctx := context.Background()
	replConn := conn.PgConn()
	log.Info("Checking replication slot", "slotname", slotName)
	var active bool
	var activePid *int
	var lsn pglogrepl.LSN
	err = conn.QueryRow(
		ctx,
		"select active, active_pid, confirmed_flush_lsn from pg_replication_slots where slot_name=$1;",
		slotName).Scan(&active, &activePid, &lsn)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, 0, fmt.Errorf("cannot check existence of replication slot: %w", err)
	}
	if active {
		return false, 0, fmt.Errorf("cannot create replication, slot is active, slot_name=%s, active_pid=%d", slotName, *activePid)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Debug("Replication slot exists", "lsn", lsn)
		return true, lsn, nil
	}
	replicationSlotResult, err := pglogrepl.CreateReplicationSlot(
		ctx,
		replConn,
		slotName, "pgoutput",
		pglogrepl.CreateReplicationSlotOptions{
			Temporary:      false,
			SnapshotAction: "",
			Mode:           pglogrepl.LogicalReplication,
		})
	if err != nil {
		return false, 0, fmt.Errorf("cannot create replication slot: %w", err)
	}
	log.Debug("Created replication slot", "result", replicationSlotResult)
	return false, xlogpos, nil
}

// pgVersion gets postgres server version and checks it is supported.
func pgVersion(log *slog.Logger, conn *pgx.Conn) (int, error) {
	log.Info("Getting server version")
	var versionString string
	err := conn.QueryRow(context.Background(), "show server_version_num;").Scan(&versionString)
	if err != nil {
		return 0, fmt.Errorf("cannot get postgres server version, error=%w", err)
	}
	version := (int(versionString[0])-48)*10 + int(versionString[1]) - 48

	log.Info("Postgres server version", "version", version)
	if version < 12 {
		return version, fmt.Errorf("unsupported postgres version=%d, minversion=12", version)
	}
	if version > 16 {
		log.Warn("New postgres version. may not be supported", "version", version, "maxversion", 16)
	}
	return version, nil
}

func DoReplicateDatabase(database SourceDatabase, url *SourceURL) {
	for {
		err := ReplicateDatabase(database, url)
		log.Error(err.Error(), "db-sid", database.Name+"_"+url.SID, "url", url.URL)
		URLError[url.URL] = err.Error()
		time.Sleep(60 * time.Second)
	}
}

//nolint:funlen,gocognit // This is is just multiple steps and needs to be in a single function
func ReplicateDatabase(database SourceDatabase, url *SourceURL) error {
	// Connect to selected source database
	log := log.With("db-sid", database.Name+"-"+url.SID)
	databaseURL := strings.Split(url.URL, "?")[0] + "?replication=database&application_name=kuvasz_" + database.Name
	parsedConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("cannot parse url=%s, error=%w", databaseURL, err)
	}
	parsedConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	dbName := parsedConfig.Database
	log.Info("Connecting", "databaseURL", databaseURL)
	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, parsedConfig)
	if err != nil {
		return fmt.Errorf("cannot start replication connection, error=%w", err)
	}
	replConn := conn.PgConn()

	// Validate server version
	log.Info("Validating server version")
	ver, err := pgVersion(log, conn)
	if err != nil {
		return fmt.Errorf("cannot get postgres server version, error=%w", err)
	}

	// Get server informatioon
	log.Info("Identifying system")
	sysident, err := pglogrepl.IdentifySystem(ctx, replConn)
	if err != nil {
		return fmt.Errorf("cannot identify system, error=%w", err)
	}
	log.Info("System identification",
		"SystemID", sysident.SystemID,
		"Timeline", sysident.Timeline,
		"XLogPos", sysident.XLogPos,
		"DBName", sysident.DBName)

	// Check for correct wal_level
	log.Info("Checking wal_level is logical")
	var walLevel string
	err = conn.QueryRow(ctx, "show wal_level").Scan(&walLevel)
	if err != nil {
		return fmt.Errorf("cannot get wal_level, error=%w", err)
	}
	if walLevel != "logical" {
		return fmt.Errorf("wal_level must be logical, got %s, change configuration in postgresql.conf and restart server", walLevel)
	}

	// Check existing publication and create if needed, drop replication slot if required
	slotName := "kuvasz_" + dbName
	slotName = strings.ReplaceAll(slotName, "-", "_")
	var publication, slot int
	err = conn.QueryRow(context.Background(), `with publication as (
							select count(*) as publication 
							from pg_publication 
							where pubname=$1), 
						slot as (select count(*) as slot 
							from pg_replication_slots 
							where slot_name=$1) 
						select publication.publication, slot.slot from publication,slot;`, slotName).Scan(&publication, &slot)
	if err != nil {
		return fmt.Errorf("cannot check publication and slot, error=%w", err)
	}
	// Check existing replication slot and existing consumer
	if !(publication == 1 && slot == 1) { // publication and slot need to be corrected
		if publication == 0 && slot == 1 { //slot without publication, drop it
			_, err = conn.Exec(context.Background(), `select pg_drop_replication_slot($1)`, slotName)
			if err != nil {
				return fmt.Errorf("cannot drop replication slot, error=%w", err)
			}
		}
		if publication == 1 && slot == 0 { // publication may have been created by mistake, remote it
			_, err = conn.Exec(context.Background(), "drop publication "+slotName)
			if err != nil {
				return fmt.Errorf("cannot drop publication, error=%w", err)
			}
		}
		_, err = conn.Exec(context.Background(), "create publication "+slotName+" for all tables")
		if err != nil {
			return fmt.Errorf("cannot create publication, error=%w", err)
		}
	}
	// Create slot if it does not exist, fail if there is an existing consumer
	oldSlot, lsn, err := createReplicationSlot(log, conn, slotName, sysident.XLogPos)
	if err != nil {
		return fmt.Errorf("cannot create replication slot, error=%w", err)
	}

	// Perform full table sync if slot was just created
	if !oldSlot {
		err := syncAllTables(log, dbName, url.SID, database.Tables, replConn)
		if err != nil {
			return fmt.Errorf("cannot perform initial sync, error=%w", err)
		}
		log.Debug("Finished full table sync")
		time.Sleep(time.Duration(config.Maintenance.StartDelay) * time.Second)
	}

	// Start replication
	log.Debug("Starting replication slot")
	protocolVersion, args := pluginArguments(ver, slotName)
	err = pglogrepl.StartReplication(
		ctx,
		replConn,
		slotName,
		lsn,
		pglogrepl.StartReplicationOptions{
			Timeline:   0,
			Mode:       pglogrepl.LogicalReplication,
			PluginArgs: args})
	if err != nil {
		return fmt.Errorf("cannot start replication, error=%w", err)
	}
	log.Info("Started logical replication slot", "slotname", slotName, "lsn", lsn)

	// Start streaming and processing messages
	clientXLogPos := sysident.XLogPos
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relations := PGRelations{}
	typeMap := pgtype.NewMap()

	// whenever we get StreamStartMessage we set inStream to true and then pass it to DecodeV2 function
	// on StreamStopMessage we set it back to false
	inStream := false

	for {
		urlHeartbeat.WithLabelValues(database.Name, url.SID).Set(float64(time.Now().Unix()))

		if time.Now().After(nextStandbyMessageDeadline) {
			err = pglogrepl.SendStandbyStatusUpdate(
				ctx,
				replConn,
				pglogrepl.StandbyStatusUpdate{
					WALWritePosition: clientXLogPos,
					WALFlushPosition: clientXLogPos,
					WALApplyPosition: clientXLogPos,
					ClientTime:       time.Now(),
					ReplyRequested:   false,
				})
			if err != nil {
				return fmt.Errorf("cannot send SendStandbyStatusUpdate, error=%w", err)
			}
			log.Debug("Sent Standby status message", "pos", clientXLogPos.String())
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		timerCtx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		rawMsg, err := replConn.ReceiveMessage(timerCtx)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) {
				continue
			}
			return fmt.Errorf("cannot ReceiveMessage, error=%w", err)
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			return fmt.Errorf("received Postgres WAL error=%v", errMsg)
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			log.Info("Received unexpected message", "message", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				log.Error("ParsePrimaryKeepaliveMessage failed", "error", err)
			}
			log.Debug("PrimaryKeepalive",
				"ServerWALEnd", pkm.ServerWALEnd,
				"ServerTime", pkm.ServerTime,
				"ReplyRequested", pkm.ReplyRequested)
			if pkm.ServerWALEnd > clientXLogPos {
				clientXLogPos = pkm.ServerWALEnd
			}
			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				return fmt.Errorf("cannot ParseXLogData, error=%w", err)
			}

			log.Debug("XLogData", "WALStart", xld.WALStart, "ServerWALEnd", xld.ServerWALEnd, "ServerTime", xld.ServerTime)
			processMessage(log, database, *url, protocolVersion, xld.WALData, relations, typeMap, &inStream)

			if xld.WALStart > clientXLogPos {
				clientXLogPos = xld.WALStart
			}
		}
	}
}
