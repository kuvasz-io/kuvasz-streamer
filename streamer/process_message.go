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
		opCode      string
		sourceTable string
		destTable   string
		id          int64
		relation    PGRelation
		values      map[string]any
		old         uint8
		oldValues   map[string]any
		lsn         pglogrepl.LSN
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

//nolint:funlen,gocognit,cyclop,gocyclo // It's OK for this to be long.
func processMessage(
	log *slog.Logger,
	database SourceDatabase,
	url SourceURL,
	version int,
	xld pglogrepl.XLogData,
	relations PGRelations,
	typeMap *pgtype.Map,
	transactionLSN *pglogrepl.LSN,
	committedTransactionLSN *pglogrepl.LSN,
	inStream *bool) {
	var sourceTable *SourceTable
	var destTable string
	var logicalMsg pglogrepl.Message
	var err error
	op := operation{
		log:      log,
		database: database.Name,
		sid:      url.SID,
		lsn:      *transactionLSN,
	}
	sourceTables := database.Tables
	walData := xld.WALData
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
		*transactionLSN = logicalMsg.FinalLSN
	case *pglogrepl.CommitMessage:
		*committedTransactionLSN = *transactionLSN
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
		sourceTable, destTable, err = sourceTables.GetTable(joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error(err.Error())
			return
		}
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		values := getValues(rel, m.Tuple.Columns, typeMap)
		op.values = values
		op.id = sourceTable.ID
		log.Debug("XLogData INSERT", "namespace", rel.Namespace, "relation", rel.RelationName, "values", values)
		if sourceTable.Type == TableTypeHistory {
			t0, _ := time.Parse("2006-01-02", "1900-01-01")
			err = op.insertHistory(destTable, t0, values)
		} else {
			op.opCode = "ic"
			SendWork(op)
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
		sourceTable, destTable, err = sourceTables.GetTable(joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error(err.Error())
			return
		}
		op.old = m.OldTupleType
		if op.old != 0 {
			op.oldValues = getValues(rel, m.OldTuple.Columns, typeMap)
		}
		op.values = getValues(rel, m.NewTuple.Columns, typeMap)
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		op.relation = rel
		op.id = sourceTable.ID
		log.Debug("XLogData UPDATE", "namespace", rel.Namespace, "relation", rel.RelationName, "oldValues", op.oldValues, "values", op.values)
		if sourceTable.Type == TableTypeHistory {
			err = op.updateHistory(destTable, rel, op.values, op.old, op.oldValues)
		} else {
			op.opCode = "uc"
			SendWork(op)
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
		sourceTable, destTable, err = sourceTables.GetTable(joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error("cannot map source table", "error", err.Error())
			return
		}
		if sourceTable.Type == TableTypeAppend {
			log.Debug("XLogDataV1 DELETE %s.%s ignored for append table type", rel.Namespace, rel.RelationName)
			return
		}
		op.values = getValues(rel, m.OldTuple.Columns, typeMap)
		log.Debug("XLogDataV1 DELETE", "namespace", rel.Namespace, "relation", rel.RelationName, "values", op.values, "old", m.OldTupleType)
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		op.relation = relations[m.RelationID]
		op.old = m.OldTupleType
		op.id = sourceTable.ID
		if sourceTable.Type == TableTypeHistory {
			err = op.deleteHistory(destTable, relations[m.RelationID], op.values, m.OldTupleType)
		} else {
			op.opCode = "dc"
			SendWork(op)
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
