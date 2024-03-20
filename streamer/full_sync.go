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
	row := slices.Clone(p)
	size += int64(len(row))
	s.rowsTotal.Inc()
	s.bytesTotal.Add(float64(len(row)))
	s.SyncDataChannel <- row
	return len(p), nil
}

func writeDestination(log *slog.Logger, tableName string, columns string, s *syncChannel) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()
	conn, err := DestConnectionPool.Acquire(ctx)
	if err != nil {
		log.Error("cannot acquire connection to destination database", "error", err)
		return
	}
	tag, err := conn.Conn().PgConn().CopyFrom(ctx, s, fmt.Sprintf("COPY %s(sid%s) FROM STDIN;", tableName, columns))
	if err != nil {
		log.Error("cannot COPY FROM", "table", tableName, "error", err)
		return
	}
	log.Debug("COPY FROM", "tag", tag)
	conn.Release()
}

func syncTable(log *slog.Logger,
	db string,
	sid string,
	sourceTableName string,
	destTableName string,
	sourceConnection *pgconn.PgConn) error {
	log = log.With("sourceTable", sourceTableName, "destTable", destTableName)
	ctx := context.Background()

	log.Debug("Starting full sync")
	// Prepare channels between reader and writer
	syncDataChannel := make(chan []byte)
	commandChannel := make(chan string)
	s := &syncChannel{
		log:             log,
		CommandChannel:  commandChannel,
		SyncDataChannel: syncDataChannel,
		rowsTotal:       syncRowsTotal.WithLabelValues("db", sid, sourceTableName),
		bytesTotal:      syncBytesTotal.WithLabelValues("db", sid, sourceTableName),
	}

	// Prepare column list
	columns := ""
	for c := range destTables[destTableName] {
		if c == "sid" || strings.HasPrefix(c, "kvsz_") {
			continue
		}
		columns = fmt.Sprintf("%s, %s", columns, c)
	}
	log.Debug("Target columns", "columns", columns)

	// Start writer
	go writeDestination(log, destTableName, columns, s)

	// Start reader
	copyStatement := fmt.Sprintf("COPY (SELECT '%s'%s FROM %s) TO STDOUT;", sid, columns, sourceTableName)
	t0 := time.Now()
	size = 0
	tag, err := sourceConnection.CopyTo(ctx, s, copyStatement)
	if err != nil {
		log.Error("cannot read source table", "error", err)
		return fmt.Errorf("cannot perform full sync, error reading source=%s, dest=%s, error=%w", sourceTableName, destTableName, err)
	}
	log.Info("Finished full sync", "tag", tag, "duration", time.Since(t0), "size", size, "throughput", (float64(size) / (time.Since(t0).Seconds()) / 1024 / 1024))

	// Stop writer
	commandChannel <- "stop"
	return nil
}

func syncAllTables(log *slog.Logger, db string, sid string, sourceTables map[string]SourceTable, sourceConnection *pgconn.PgConn) error {
	log.Debug("Sync all tables", "sourceTables", sourceTables)
	for sourceTableName := range sourceTables {
		_, destTableName, err := MapSourceTable(sourceTableName, sourceTables)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		log.Debug("Syncing", "sourceTable", sourceTableName, "destTable", destTableName)
		_ = syncTable(log, db, sid, sourceTableName, destTableName, sourceConnection)
	}
	return nil
}
