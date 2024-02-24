package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgtype"
)

type (
	operation struct {
		log         *slog.Logger
		database    string
		sid         string
		sourceTable string
	}
)

func decodeTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (any, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		log.Debug("found", "data", data, "dt", dt)
		decodedColumn, err := dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
		if err != nil {
			return decodedColumn, fmt.Errorf("cannot decode text column data %v, error: %w", data, err)
		}
		return decodedColumn, nil
	}
	log.Debug("not found", "data", data)
	return string(data), nil
}

// func decodeBinaryColumnData(mi *pgtype.Map, data []byte, dataType uint32) (any, error) {
// 	if dt, ok := mi.TypeForOID(dataType); ok {
// 		log.Debug("found", "data", data, "dt", dt)
// 		decodedColumn, err := dt.Codec.DecodeValue(mi, dataType, pgtype.BinaryFormatCode, data)
// 		if err != nil {
// 			return decodedColumn, fmt.Errorf("cannot decode text column data %v, error: %w", data, err)
// 		}
// 		return decodedColumn, nil
// 	}
// 	log.Debug("not found", "data", data)
// 	return string(data), nil
// }

func getValues(rel PGRelation, columns []*pglogrepl.TupleDataColumn, typeMap *pgtype.Map) map[string]any {
	values := map[string]any{}
	for idx, col := range columns {
		colName := rel.Columns[idx].Name
		switch col.DataType {
		case 'n': // null
			values[colName] = nil
		case 'u': // unchanged toast
			// This TOAST value was not changed. TOAST values are not stored in the tuple,
			// and logical replication doesn't want to spend a disk read to fetch its value for you.
		case 't': // text
			val, err := decodeTextColumnData(typeMap, col.Data, rel.Columns[idx].DataTypeOID)
			if err != nil {
				log.Error("error decoding column data", "error", err)
				continue
			}
			values[colName] = val
		}
	}
	return values
}

//nolint:funlen,gocognit,gocyclo,cyclop // It's OK for this to be long.
func processMessage(
	log *slog.Logger,
	database SourceDatabase,
	url SourceURL,
	version int,
	walData []byte,
	relations PGRelations,
	typeMap *pgtype.Map,
	inStream *bool) {
	var destTable string
	var logicalMsg pglogrepl.Message
	var err error
	op := operation{
		log:      log,
		database: database.Name,
		sid:      url.SID,
	}
	sourceTables := database.Tables
	switch version {
	case 1:
		logicalMsg, err = pglogrepl.Parse(walData)
	case 2:
		logicalMsg, err = pglogrepl.ParseV2(walData, *inStream)
	default:
		log.Error("Unsupported message version", "version", version)
	}
	if err != nil {
		log.Error("Parse logical replication message", "error", err)
		return
	}
	log.Debug("XLogData", "version", version, "type", logicalMsg.Type(), "message", logicalMsg)
	switch logicalMsg := logicalMsg.(type) {
	case *pglogrepl.BeginMessage:
	case *pglogrepl.CommitMessage:
	case *pglogrepl.TypeMessage:
	case *pglogrepl.OriginMessage:
	case *pglogrepl.LogicalDecodingMessage:
		log.Debug("Logical decoding message", "prefix", logicalMsg.Prefix, "content", logicalMsg.Content)

	case *pglogrepl.RelationMessage, *pglogrepl.RelationMessageV2:
		var m *pglogrepl.RelationMessage
		if version == 1 {
			m, _ = logicalMsg.(*pglogrepl.RelationMessage)
		} else {
			temp, _ := logicalMsg.(*pglogrepl.RelationMessageV2)
			m = &temp.RelationMessage
		}
		log.Debug("debug", "logicalMsg", logicalMsg, "m", m)
		rel := PGRelation{
			Namespace:    m.Namespace,
			RelationName: m.RelationName,
			Columns:      make([]PGColumn, 0),
		}
		for _, c := range m.Columns {
			rel.Columns = append(rel.Columns, PGColumn{
				Name:        c.Name,
				PrimaryKey:  (c.Flags > 0),
				DataTypeOID: c.DataType,
				ColumnType:  "",
			})
		}
		relations[m.RelationID] = rel

	case *pglogrepl.InsertMessage, *pglogrepl.InsertMessageV2:
		var m *pglogrepl.InsertMessage
		if version == 1 {
			m, _ = logicalMsg.(*pglogrepl.InsertMessage)
		} else {
			temp, _ := logicalMsg.(*pglogrepl.InsertMessageV2)
			m = &temp.InsertMessage
		}

		rel, ok := relations[m.RelationID]
		if !ok {
			log.Error("unknown relation, protocol bug", "ID", m.RelationID)
			return
		}
		sourceTable, destTable, err := MapSourceTable(rel.RelationName, sourceTables)
		if err != nil {
			log.Error(err.Error())
			return
		}
		op.sourceTable = rel.RelationName
		values := getValues(rel, m.Tuple.Columns, typeMap)
		log.Debug(fmt.Sprintf("XLogData INSERT %s.%s: %v", rel.Namespace, rel.RelationName, values))
		if sourceTable.Type == TableTypeHistory {
			t0, _ := time.Parse("2006-01-02", "1900-01-01")
			err = op.insertHistory(destTable, t0, values)
		} else {
			err = op.insertClone(destTable, values)
		}
		if err != nil {
			return
		}

	case *pglogrepl.UpdateMessage, *pglogrepl.UpdateMessageV2:
		var m *pglogrepl.UpdateMessage
		if version == 1 {
			m, _ = logicalMsg.(*pglogrepl.UpdateMessage)
		} else {
			temp, _ := logicalMsg.(*pglogrepl.UpdateMessageV2)
			m = &temp.UpdateMessage
		}

		rel, ok := relations[m.RelationID]
		if !ok {
			log.Error("unknown relation, protocol bug", "ID", m.RelationID)
			return
		}
		_, destTable, err = MapSourceTable(rel.RelationName, sourceTables)
		if err != nil {
			log.Error(err.Error())
			return
		}
		oldValues := map[string]any{}
		old := m.OldTupleType
		if old != 0 {
			oldValues = getValues(rel, m.OldTuple.Columns, typeMap)
		}
		newValues := getValues(rel, m.NewTuple.Columns, typeMap)
		log.Debug(fmt.Sprintf("XLogData UPDATE %s.%s: %v -> %v", rel.Namespace, rel.RelationName, oldValues, newValues))
		op.sourceTable = rel.RelationName
		if sourceTables[rel.RelationName].Type == TableTypeHistory {
			err = op.updateHistory(destTable, relations[m.RelationID], newValues, old, oldValues)
		} else {
			err = op.updateClone(destTable, relations[m.RelationID], newValues, old, oldValues)
		}
		if err != nil {
			return
		}

	case *pglogrepl.DeleteMessage, *pglogrepl.DeleteMessageV2:
		var m *pglogrepl.DeleteMessage
		if version == 1 {
			m, _ = logicalMsg.(*pglogrepl.DeleteMessage)
		} else {
			temp, _ := logicalMsg.(*pglogrepl.DeleteMessageV2)
			m = &temp.DeleteMessage
		}

		rel, ok := relations[m.RelationID]
		if !ok {
			log.Error("unknown relation, protocol bug", "ID", m.RelationID)
			return
		}
		sourceTable, destTable, err := MapSourceTable(rel.RelationName, sourceTables)
		if err != nil {
			log.Error(err.Error())
			return
		}
		if sourceTable.Type == TableTypeAppend {
			log.Debug("XLogDataV1 DELETE %s.%s ignored for append table type", rel.Namespace, rel.RelationName)
			return
		}
		values := getValues(rel, m.OldTuple.Columns, typeMap)
		log.Debug(fmt.Sprintf("XLogDataV1 DELETE %s.%s: %v, old: %c", rel.Namespace, rel.RelationName, values, m.OldTupleType))
		op.sourceTable = rel.RelationName
		if sourceTable.Type == TableTypeHistory {
			err = op.deleteHistory(destTable, relations[m.RelationID], values, m.OldTupleType)
		} else {
			err = op.deleteClone(destTable, relations[m.RelationID], values, m.OldTupleType)
		}
		if err != nil {
			return
		}

	case *pglogrepl.TruncateMessage:
	case *pglogrepl.TruncateMessageV2:
		// ...

	default:
		log.Warn(fmt.Sprintf("Unknown message type in pgoutput stream: %T", logicalMsg))
	}
}
