package main

import (
	"context"
	"errors"
	"fmt"
)

type (
	arg struct {
		Attribute string
		Value     any
	}
)

func (op operation) buildSetList(tableName string, args []arg, values map[string]any) ([]arg, error) {
	i := 0
	for attribute, value := range values {
		_, ok := destTables[tableName][attribute]
		if !ok {
			log.Debug("skip non-existing attribute", "attribute", attribute)
			continue
		}
		log.Debug("Add", "attribute", attribute, "value", value)
		args = append(args, arg{attribute, value})
		i++
	}
	// if no column found return error
	if i == 0 {
		return args, errors.New("no attributes were mapped")
	}
	return args, nil
}

func (op operation) buildWhere( //nolint:gocognit  // ignore complexity linter
	tableName string,
	relation PGRelation,
	values map[string]any,
	oldValues map[string]any,
	old uint8,
	query string,
	queryParameters []any) (string, []any) {
	j := len(queryParameters) + 1

	switch old {
	case 'K', 0:
		for _, column := range relation.Columns {
			if !column.PrimaryKey {
				log.Debug("Skip non-primary key component", "column", column)
				continue
			}
			_, ok := destTables[tableName][column.Name]
			if !ok {
				log.Error("Configuration error: primary key component does not exist in destination table.", "column", column)
				continue
			}
			var value any
			if old == 'K' {
				value, ok = oldValues[column.Name]
			} else {
				value, ok = values[column.Name]
			}
			if !ok {
				log.Error("Bug: Primary key component not received", "column", column)
				continue
			}
			if value == nil {
				log.Error("Bug: NULL received in primary component", "column", column)
				continue
			}
			query = fmt.Sprintf("%s AND %s=$%d", query, column.Name, j)
			queryParameters = append(queryParameters, value)
			j++
		}
	case 'O':
		// no primary key is defined, range over all incoming values skipping non-existing columns
		for column, value := range oldValues {
			_, ok := destTables[tableName][column]
			if !ok {
				log.Error("Configuration error: skip non-existing column with replica-identity-full, tables should be similar",
					"column", column)
				continue
			}
			log.Debug("Add", "column", column, "value", value)
			if value == nil {
				query = fmt.Sprintf("%s AND %s IS NULL", query, column)
			} else {
				query = fmt.Sprintf("%s AND %s=$%d", query, column, j)
				queryParameters = append(queryParameters, value)
				j++
			}
		}
	default:
		log.Error("Invalid old tuple indicator", "old", old)
	}
	return query, queryParameters
}

func (op operation) insertClone(tableName string, values map[string]any) error {
	var query string
	log = op.log.With("op", "insertClone", "table", tableName)

	// Build query
	columns := "sid"
	valuesIndices := "$1"
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, op.sid)
	i := 2
	for c, v := range values {
		_, ok := destTables[tableName][c]
		if !ok {
			log.Debug("skip non-existing destination column", "column", c)
			continue
		}
		log.Debug("Add", "column", c, "value", v)
		columns = fmt.Sprintf("%s, %s", columns, c)
		valuesIndices = fmt.Sprintf("%s, $%d", valuesIndices, i)
		queryParameters = append(queryParameters, v)
		i++
	}
	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s) on conflict do nothing", tableName, columns, valuesIndices)

	// Run query
	log.Debug("insert", "query", query)
	_, err = DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		log.Error("can't insert", "table", tableName, "query", query, "error", err)
		requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "insert", "failure").Inc()
		return fmt.Errorf("insertClone failed, error=%w", err)
	}
	requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "insert", "success").Inc()
	return nil
}

// Cases
// 1. PK exists and is not updated => old = 0, oldValues=nil ==> where PK=PK and sid=SID.
// 2. PK exists and is updated => old=K, oldValues=oldPK ==> where PK=oldPK and sid=SID.
// 3. PK does not exist, replica full => old=O, oldValues=alloldValues ==> where allfields=alloldValues.
func (op operation) updateClone(tableName string, relation PGRelation, values map[string]any, old uint8, oldValues map[string]any) error {
	var i int
	args := make([]arg, 0)
	log = op.log.With("op", "updateClone", "table", tableName)

	log.Debug("Dump params", "values", values, "oldvalues", oldValues, "old", old)
	// Build argument list
	args = append(args, arg{"sid", op.sid})
	args, err := op.buildSetList(tableName, args, values)
	if err != nil {
		return err
	}

	// Start building UPDATE query
	query := fmt.Sprintf("UPDATE %s SET %s=$1", tableName, args[0].Attribute)
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, args[0].Value)
	for i = 1; i < len(args); i++ {
		query = fmt.Sprintf("%s, %s=$%d", query, args[i].Attribute, i+1)
		queryParameters = append(queryParameters, args[i].Value)
	}

	// Add WHERE clause
	query = fmt.Sprintf("%s WHERE sid=$%d", query, i+1)
	queryParameters = append(queryParameters, op.sid)

	// add primary key
	query, queryParameters = op.buildWhere(tableName, relation, values, oldValues, old, query, queryParameters)

	// Run query
	log.Debug("update", "query", query, "queryParameters", queryParameters)
	_, err = DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		log.Error("can't update", "table", tableName, "query", query, "error", err)
		requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "update", "failure").Inc()
		return fmt.Errorf("updateClone failed: error=%w", err)
	}
	// requestDuration.WithLabelValues(path, r.Method, code).Observe(float64(duration) / 1000)
	requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "update", "success").Inc()

	return nil
}

func (op operation) deleteClone(tableName string, relation PGRelation, values map[string]any, old uint8) error {
	var query string
	log = op.log.With("op", "deleteClone", "table", tableName)

	log.Debug("Dump params", "relation", relation, "values", values, "old", old)
	// Build query
	query = fmt.Sprintf("DELETE FROM %s WHERE sid=$1 ", tableName)
	queryParameters := make([]any, 0)
	queryParameters = append(queryParameters, op.sid)

	query, queryParameters = op.buildWhere(tableName, relation, nil, values, old, query, queryParameters)
	// Run query
	log.Debug("delete", "query", query, "parameters", queryParameters)
	rows, err := DestConnectionPool.Exec(context.Background(), query, queryParameters...)
	if err != nil {
		requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "delete", "failure").Inc()
		log.Error("can't delete", "table", tableName, "query", query, "error", err)
		return fmt.Errorf("deleteClone failed: error=%w", err)
	}
	if rows.RowsAffected() == 0 {
		log.Error("did not find row to delete, destination database was not in sync", "query", query, "parameters", queryParameters)
		requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "delete", "failure").Inc()
		return fmt.Errorf("deleteClone failed: no affected rows")
	}
	requestsTotal.WithLabelValues(op.database, op.sid, op.sourceTable, "delete", "success").Inc()
	return nil
}
