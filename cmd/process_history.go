package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func insertHistory(log *slog.Logger, sid string, tableName string, startTime time.Time, values map[string]any) error {
	var query string
	args := make([]arg, 0)
	log = log.With("op", "insertHistory", "table", tableName)

	// Build argument list
	args = append(args, arg{"sid", sid})
	args = append(args, arg{"kvsz_start", startTime})
	args = append(args, arg{"kvsz_end", "9999-01-01 00:00:00"})
	args = append(args, arg{"kvsz_deleted", false})
	args, err := buildSetList(log, tableName, args, values)
	if err != nil {
		return err
	}
	log.Debug("Dump params", "values", values)

	// Build query
	attributes := args[0].Attribute
	valuesIndices := "$1"
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, args[0].Value)
	for i := 1; i < len(args); i++ {
		attributes = fmt.Sprintf("%s, %s", attributes, args[i].Attribute)
		valuesIndices = fmt.Sprintf("%s, $%d", valuesIndices, i+1)
		queryParameters = append(queryParameters, args[i].Value)
	}
	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", tableName, attributes, valuesIndices)

	// Run query
	log.Debug("insert", "query", query)
	_, err = DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		log.Error("can't insert", "query", query, "error", err)
		return fmt.Errorf("insertHistory failed, error=%w", err)
	}
	return nil
}

// Cases
// 1. PK exists and is not updated => old = 0, oldValues=nil ==> where PK=PK and sid=SID.
// 2. PK exists and is updated => old=K, oldValues=oldPK ==> where PK=oldPK and sid=SID.
// 3. PK does not exist, replica full => old=O, oldValues=alloldValues ==> where allfields=alloldValues.
func updateHistory(log *slog.Logger, sid string, tableName string, relation PGRelation,
	values map[string]any, old uint8, oldValues map[string]any) error {
	var i = 1
	log = log.With("op", "updateHistory", "table", tableName)

	t0 := time.Now()

	log.Debug("Dump params", "values", values, "oldvalues", oldValues, "old", old)

	// Update old record with kvsz_end=now
	query := fmt.Sprintf("UPDATE %s SET kvsz_end=$1", tableName)
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, t0)
	i++

	// Add WHERE clause
	query = fmt.Sprintf("%s WHERE sid=$%d AND kvsz_end='9999-01-01'", query, i)
	queryParameters = append(queryParameters, sid)
	query, queryParameters = buildWhere(log, tableName, relation, nil, values, old, query, queryParameters)

	// Run query
	log.Debug("update", "query", query, "args", queryParameters)
	_, err = DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		log.Error("can't update", "table", tableName, "query", query, "error", err)
		return fmt.Errorf("updateHistory failed: error=%w", err)
	}
	err = insertHistory(log, sid, tableName, t0, values)
	return nil
}

func deleteHistory(log *slog.Logger, sid string, tableName string, relation PGRelation, values map[string]any, old uint8) error {
	var query string
	log = log.With("op", "deleteHistory", "table", tableName)
	t0 := time.Now()

	// Build query
	query = fmt.Sprintf("UPDATE %s set kvsz_deleted=true, kvsz_end=$2 WHERE sid=$1 AND kvsz_end='9999-01-01' ", tableName)
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, sid)
	queryParameters = append(queryParameters, t0)

	query, queryParameters = buildWhere(log, tableName, relation, nil, values, old, query, queryParameters)
	// Run query
	log.Debug("delete",
		"query", query,
		"queryParameters", queryParameters)
	rows, err := DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		log.Error("can't update history table",
			"query", query,
			"queryParameters", queryParameters,
			"error", err)
		return fmt.Errorf("deleteHistory failed: error=%w", err)
	}
	if rows.RowsAffected() == 0 {
		log.Error("did not find row to delete, destination database was not in sync",
			"query", query,
			"queryParameters", queryParameters)
		return fmt.Errorf("deleteHistory failed: no affected rows")
	}
	log.Debug("delete", "RowsAffected", rows.RowsAffected())
	return nil
}
