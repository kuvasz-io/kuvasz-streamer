package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type (
	syncReader struct {
		DataChannel    chan []byte
		CommandChannel chan string
	}
	syncWriter struct {
		DataChannel chan []byte
	}
)

func (r syncReader) Read(p []byte) (int, error) {
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
	log.Debug("write chunk", "chunk", p)
	w.DataChannel <- p
	return len(p), nil
}

func writeDestination(log *slog.Logger, tableName string, columns string, dataChannel chan []byte, commandChannel chan string) {
	r := &syncReader{DataChannel: dataChannel, CommandChannel: commandChannel}
	ctx := context.Background()
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
}

func syncTable(log *slog.Logger,
	sid string,
	sourceTableName string,
	destTableName string,
	sourceConnection *pgconn.PgConn) error {
	ctx := context.Background()
	// Prepare channels between reader and writer
	dataChannel := make(chan []byte)
	commandChannel := make(chan string)
	w := &syncWriter{DataChannel: dataChannel}

	// Prepare column list
	columns := ""
	for c := range destTables[destTableName] {
		if c == "sid" || strings.HasPrefix(c, "kvsz_") {
			continue
		}
		columns = fmt.Sprintf("%s, %s", columns, c)
	}

	// Start writer
	go writeDestination(log, destTableName, columns, dataChannel, commandChannel)

	// Start reader
	copyStatement := fmt.Sprintf("COPY (SELECT '%s'%s FROM %s) TO STDOUT;", sid, columns, sourceTableName)
	tag, err := sourceConnection.CopyTo(ctx, w, copyStatement)
	if err != nil {
		log.Error("Cannot copy t1", "error", err)
		return fmt.Errorf("cannot perform full sync")
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
