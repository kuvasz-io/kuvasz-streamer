package main

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgtype"
)

type (
	operation struct {
		log             *slog.Logger
		database        string
		sid             string
		opCode          string
		sourceTable     string
		destTable       string
		destTableHasSID bool
		id              int64
		relation        PGRelation
		values          map[string]any
		old             uint8
		oldValues       map[string]any
		lsn             pglogrepl.LSN
	}
)

func decodeTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (any, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		log.Debug("found", "data", string(data), "dt", dt)
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
	log.Debug("got values", "relation", rel, "values", values)
	return values
}

func getEnv(rel PGRelation, columns []*pglogrepl.TupleDataColumn, typeMap *pgtype.Map) map[string]any {
	values := map[string]any{}
	for idx, col := range columns {
		colName := rel.Columns[idx].Name
		if colName == "type" {
			colName = "_type"
		}
		switch col.DataType {
		case 'n': // null
			values[colName] = nil
		case 'u': // unchanged toast
			// This TOAST value was not changed. TOAST values are not stored in the tuple,
			// and logical replication doesn't want to spend a disk read to fetch its value for you.
		case 't': // text
			if rel.Columns[idx].DataTypeOID == 2950 { // uuid
				values[colName] = string(col.Data)
				continue
			}
			if dt, ok := typeMap.TypeForOID(rel.Columns[idx].DataTypeOID); ok {
				log.Debug("found", "data", string(col.Data), "dt", dt)
				decodedColumn, err := dt.Codec.DecodeValue(typeMap, rel.Columns[idx].DataTypeOID, pgtype.TextFormatCode, col.Data)
				if err != nil {
					log.Error("cannot decode text column", "data", col.Data, "error", err)
					continue
				}
				values[colName] = decodedColumn
				continue
			}
			values[colName] = string(col.Data)
		}
	}
	return values
}

func filter(log *slog.Logger, filterExpression cel.Program, args map[string]any) bool {
	// no filter defined, pass
	if filterExpression == nil {
		return true
	}
	result, err := evalExpression(filterExpression, args)
	// filter error, allow -- maybe we should take a default from config
	if err != nil {
		log.Error("cannot run filter, allowing", "args", args, "error", err)
		return true
	}
	log.Debug("applied filter", "args", args, "result", result, "type", reflect.TypeOf(result))
	switch t := result.(type) {
	case types.Bool:
		return bool(t)
	default:
		log.Error("filter did not return boolean, allowing")
	}
	return true
}

func setter(log *slog.Logger, setExpression cel.Program, args map[string]any) any {
	result, err := evalExpression(setExpression, args)
	if err != nil {
		log.Error("cannot run setter", "error", err)
		return nil
	}
	log.Debug("setter", "args", args, "output", result, "type", reflect.TypeOf(result))
	return result
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
	var logicalMsg pglogrepl.Message
	var err error
	op := operation{
		log:      log,
		database: database.Name,
		sid:      url.SID,
		lsn:      *transactionLSN,
	}
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
		entry, err := MappingTable.FindByName(database.Name, joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error("cannot match table", "schema", rel.Namespace, "table", rel.RelationName)
			return
		}
		values := getValues(rel, m.Tuple.Columns, typeMap)
		env := getEnv(rel, m.Tuple.Columns, typeMap)
		if !filter(log, entry.compiledFilter, env) {
			return
		}
		if len(entry.compiledSet) != 0 { // we have set, map values to new ones
			setValues := make(map[string]any)
			for v, p := range entry.compiledSet {
				setValues[v] = setter(log, p, env)
			}
			values = setValues
		}

		log.Debug("XLogData INSERT", "namespace", rel.Namespace, "relation", rel.RelationName, "values", values)
		destTable := joinSchema(config.Database.Schema, entry.Target)
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		_, op.destTableHasSID = DestTables[destTable].Columns["sid"]
		op.values = values
		op.id = entry.ID
		if entry.Type == TableTypeHistory {
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
		entry, err := MappingTable.FindByName(database.Name, joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error("cannot match table", "schema", rel.Namespace, "table", rel.RelationName)
			return
		}
		op.old = m.OldTupleType
		if op.old != 0 {
			op.oldValues = getValues(rel, m.OldTuple.Columns, typeMap)
		}
		values := getValues(rel, m.NewTuple.Columns, typeMap)
		env := getEnv(rel, m.NewTuple.Columns, typeMap)
		if !filter(log, entry.compiledFilter, env) {
			return
		}
		if len(entry.compiledSet) != 0 { // we have set, map values to new ones
			setValues := make(map[string]any)
			for v, p := range entry.compiledSet {
				setValues[v] = setter(log, p, env)
			}
			values = setValues
		}
		op.values = values

		log.Debug("XLogData UPDATE", "namespace", rel.Namespace, "relation", rel.RelationName, "oldValues", op.oldValues, "values", op.values)
		destTable := joinSchema(config.Database.Schema, entry.Target)
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		_, op.destTableHasSID = DestTables[destTable].Columns["sid"]
		op.relation = rel
		op.id = entry.ID
		if entry.Type == TableTypeHistory {
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
		entry, err := MappingTable.FindByName(database.Name, joinSchema(rel.Namespace, rel.RelationName))
		if err != nil {
			log.Error("cannot match table", "schema", rel.Namespace, "table", rel.RelationName)
			return
		}
		if entry.Type == TableTypeAppend {
			log.Debug("XLogDataV1 DELETE %s.%s ignored for append table type", rel.Namespace, rel.RelationName)
			return
		}
		op.values = getValues(rel, m.OldTuple.Columns, typeMap)
		env := getEnv(rel, m.OldTuple.Columns, typeMap)
		if !filter(log, entry.compiledFilter, env) {
			return
		}
		log.Debug("XLogDataV1 DELETE", "namespace", rel.Namespace, "relation", rel.RelationName, "values", op.values, "old", m.OldTupleType)
		destTable := joinSchema(config.Database.Schema, entry.Target)
		op.sourceTable = rel.RelationName
		op.destTable = destTable
		_, op.destTableHasSID = DestTables[destTable].Columns["sid"]
		op.relation = relations[m.RelationID]
		op.old = m.OldTupleType
		op.id = entry.ID
		if entry.Type == TableTypeHistory {
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
		log.Warn("Unknown message type in pgoutput stream", "type", logicalMsg.Type().String())
	}
}
