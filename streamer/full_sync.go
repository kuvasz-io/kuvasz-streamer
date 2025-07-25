package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	syncChannel struct {
		log             *slog.Logger
		SyncDataChannel chan []byte
		CommandChannel  chan string
		rowsTotal       prometheus.Counter
		bytesTotal      prometheus.Counter
	}
)

var size int64

func (s syncChannel) Read(p []byte) (int, error) {
	select {
	case command := <-s.CommandChannel:
		log.Debug("received command", "command", command)
		return 0, io.EOF
	case row := <-s.SyncDataChannel:
		n := copy(p, row)
		return n, nil
	}
}

func (s syncChannel) Write(p []byte) (int, error) {
	err := lim.Wait(context.Background())
	if err != nil {
		return 0, fmt.Errorf("cannot wait for token, error=%w", err)
	}
	row := slices.Clone(p)
	size += int64(len(row))
	s.rowsTotal.Inc()
	s.bytesTotal.Add(float64(len(row)))
	s.SyncDataChannel <- row
	return len(p), nil
}

func writeDestination(log *slog.Logger, tableName string, hasSID bool, columns string, s *syncChannel) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()
	conn, err := DestConnectionPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		log.Error("cannot acquire connection to destination database", "error", err)
		return
	}
	var tag pgconn.CommandTag
	if hasSID {
		tag, err = conn.Conn().PgConn().CopyFrom(ctx, s, fmt.Sprintf("COPY %s(sid, %s) FROM STDIN;", tableName, columns))
	} else {
		tag, err = conn.Conn().PgConn().CopyFrom(ctx, s, fmt.Sprintf("COPY %s(%s) FROM STDIN;", tableName, columns))
	}
	if err != nil {
		log.Error("cannot COPY FROM", "table", tableName, "error", err)
		return
	}
	log.Debug("COPY FROM", "tag", tag)
}

func syncTable(log *slog.Logger,
	db string,
	sid string,
	sourceTableName string,
	destTableName string,
	sourceConnection *pgconn.PgConn) error {
	log = log.With("sourceTable", sourceTableName, "destTable", destTableName)
	ctx := context.Background()
	hasSID := false

	log.Debug("Starting full sync")
	// Prepare channels between reader and writer
	syncDataChannel := make(chan []byte)
	commandChannel := make(chan string)
	s := &syncChannel{
		log:             log,
		CommandChannel:  commandChannel,
		SyncDataChannel: syncDataChannel,
		rowsTotal:       syncRowsTotal.WithLabelValues(db, sid, sourceTableName),
		bytesTotal:      syncBytesTotal.WithLabelValues(db, sid, sourceTableName),
	}

	// Find map entry for source table
	mapentry, err := MappingTable.FindByName(db, sourceTableName)
	if err != nil {
		log.Error("cannot match table", "database", db, "table", sourceTableName)
		return fmt.Errorf("cannot find table: %s", sourceTableName)
	}
	log.Debug("Found mapping entry", "map", MappingTable, "mapentry", mapentry)

	// Prepare column list
	columns := ""
	for c := range DestTables[destTableName].Columns {
		if strings.HasPrefix(c, "kvsz_") {
			continue
		}
		if c == "sid" {
			hasSID = true
			continue
		}
		if _, ok := mapentry.SourceColumns[c]; !ok {
			log.Debug("Target column not found in source table", "column", c, "mapentry.SourceColumns[c]", mapentry.SourceColumns[c])
			continue
		}
		if columns == "" {
			columns = c
		} else {
			columns = fmt.Sprintf("%s, %s", columns, c)
		}
	}
	log.Debug("Target columns", "columns", columns)

	// Start writer
	go writeDestination(log, destTableName, hasSID, columns, s)

	// Start reader
	var copyStatement string
	if hasSID {
		copyStatement = fmt.Sprintf("COPY (SELECT '%s', %s FROM %s) TO STDOUT;", sid, columns, sourceTableName)
	} else {
		copyStatement = fmt.Sprintf("COPY (SELECT %s FROM %s) TO STDOUT;", columns, sourceTableName)
	}
	t0 := time.Now()
	size = 0
	tag, err := sourceConnection.CopyTo(ctx, s, copyStatement)
	if err != nil {
		log.Error("cannot read source table", "error", err)
		return fmt.Errorf("cannot perform full sync, error reading source=%s, dest=%s, error=%w", sourceTableName, destTableName, err)
	}
	log.Info("Finished full sync",
		"tag", tag,
		"duration", time.Since(t0), "size",
		size, "throughput",
		(float64(size) / (time.Since(t0).Seconds()) / 1024 / 1024))

	// Stop writer
	commandChannel <- "stop"
	return nil
}

func syncAllTables(
	log *slog.Logger,
	db string,
	sid string,
	sourceTables SourceTables,
	sourceConnection *pgconn.PgConn) error {
	log.Info("Starting full sync for all tables", "sourceTables", sourceTables)
	for sourceTableName := range sourceTables {
		destTableName, err := sourceTables.GetTable(sourceTableName)
		if err != nil {
			return err
		}
		log.Info("Syncing", "sourceTable", sourceTableName, "destTable", destTableName)
		_ = syncTable(log, db, sid, sourceTableName, destTableName, sourceConnection)
	}
	return nil
}

func syncNewTables(
	log *slog.Logger,
	db string,
	sid string,
	sourceTables SourceTables,
	newTables []string,
	sourceConnection *pgconn.PgConn) error {
	log.Info("Starting full sync for new tables", "sourceTables", sourceTables)
	for i := range newTables {
		destTableName, err := sourceTables.GetTable(newTables[i])
		if err != nil {
			return err
		}
		log.Info("Syncing", "sourceTable", newTables[i], "destTable", destTableName)
		_ = syncTable(log, db, sid, newTables[i], destTableName, sourceConnection)
	}
	return nil
}
