package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type (
	syncReader struct {
		DataChannel    chan []byte
		CommandChannel chan string
		log            *slog.Logger
	}
	syncWriter struct {
		DataChannel chan []byte
		log         *slog.Logger
	}
)

func (r syncReader) Read(p []byte) (int, error) {
	log := r.log
	select {
	case command := <-r.CommandChannel:
		log.Debug("received command", "command", command)
		return 0, io.EOF
	case row := <-r.DataChannel:
		log.Debug("read row", "row", row)
		n := copy(p, row)
		return n, nil
	}
}

func (w syncWriter) Write(p []byte) (int, error) {
	log := w.log
	log.Debug("write chunk", "chunk", p)
	w.DataChannel <- p
	return len(p), nil
}

func writeDestination(log *slog.Logger, tableName string, columns string, dataChannel chan []byte, commandChannel chan string) {
	r := &syncReader{DataChannel: dataChannel, CommandChannel: commandChannel, log: log}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	conn, err := DestConnectionPool.Acquire(ctx)
	if err != nil {
		log.Error("cannot acquire connection to destination database", "error", err)
		return
	}
	tag, err := conn.Conn().PgConn().CopyFrom(ctx, r, fmt.Sprintf("COPY %s(sid%s) FROM STDIN;", tableName, columns))
	if err != nil {
		log.Error("cannot COPY FROM", "table", "error", tableName, err)
		return
	}
	log.Debug("COPY FROM", "tag", tag)
	conn.Release()
}

func syncTable(log *slog.Logger,
	sid string,
	sourceTableName string,
	destTableName string,
	sourceConnection *pgconn.PgConn) error {
	log = log.With("sourceTable", sourceTableName, "destTable", destTableName)
	ctx := context.Background()

	log.Debug("Starting full sync")
	// Prepare channels between reader and writer
	dataChannel := make(chan []byte)
	commandChannel := make(chan string)
	w := &syncWriter{DataChannel: dataChannel, log: log}

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
	go writeDestination(log, destTableName, columns, dataChannel, commandChannel)

	// Start reader
	copyStatement := fmt.Sprintf("COPY (SELECT '%s'%s FROM %s) TO STDOUT;", sid, columns, sourceTableName)
	tag, err := sourceConnection.CopyTo(ctx, w, copyStatement)
	if err != nil {
		log.Error("Cannot read source table", "error", err)
		return fmt.Errorf("cannot perform full sync, error reading source=%s, dest=%s", sourceTableName, destTableName)
	}
	log.Debug("COPY TO", "tag", tag)

	// Stop writer
	commandChannel <- "stop"
	return nil
}

func syncAllTables(log *slog.Logger, sid string, sourceTables map[string]SourceTable, sourceConnection *pgconn.PgConn) error {
	for sourceTableName := range sourceTables {
		destTableName, err := MapSourceTable(sourceTableName, sourceTables)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		_ = syncTable(log, sid, sourceTableName, destTableName, sourceConnection)
	}
	return nil
}
