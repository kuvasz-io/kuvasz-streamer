package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	PGColumn struct {
		Name        string
		ColumnType  string
		DataTypeOID uint32
		PrimaryKey  bool
	}
	PGTable  map[string]PGColumn
	PGTables map[string]PGTable

	PGRelation struct {
		Namespace    string
		RelationName string
		Columns      []PGColumn
	}
	PGRelations map[uint32]PGRelation
)

// func getPrimaryKey(log *slog.Logger, database *pgxpool.Pool, schemaName string, tableName string) (map[string]bool, error) {
// 	result := make(map[string]bool)
// 	query := `SELECT c.column_name FROM information_schema.table_constraints tc
// 			  JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
// 			  JOIN information_schema.columns AS c
// 			    ON c.table_schema = tc.constraint_schema  AND tc.table_name = c.table_name AND ccu.column_name = c.column_name
// 			  WHERE constraint_type = 'PRIMARY KEY'
// 			  	and tc.constraint_catalog =current_database()
// 				and c.table_schema = $1
// 				and tc.table_name = $2`
// 	pkRows, err := database.Query(context.Background(), query, schemaName, tableName)
// 	if err != nil {
// 		log.Error("cannot get contraints", "error", err)
// 		return result, fmt.Errorf("cannot get contraints, table=%s, error=%w", tableName, err)
// 	}
// 	defer pkRows.Close()

// 	for pkRows.Next() {
// 		var columnName string

// 		err = pkRows.Scan(&columnName)
// 		if err != nil {
// 			return result, fmt.Errorf("cannot map row constraints to values, table=%s, column=%s, error=%w", tableName, columnName, err)
// 		}
// 		result[columnName] = true
// 	}
// 	return result, nil
// }

func GetTables(log *slog.Logger, database *pgxpool.Pool, schemaName string) (PGTables, error) {
	query := `SELECT c.table_name, c.column_name, c.udt_name, t.oid from information_schema.columns as c
			  INNER JOIN pg_type as t ON c.udt_name=t.typname
	          WHERE c.table_catalog=current_database() and c.table_schema=$1;`

	pgTables := make(PGTables)
	log = log.With("schema", schemaName)
	log.Debug("Fetching tables and columns")
	rows, err := database.Query(context.Background(), query, schemaName)
	if err != nil {
		return pgTables, fmt.Errorf("cannot get column metadata from destination database, schema=%s, error=%w", schemaName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var pgColumn PGColumn
		var columnName string
		err = rows.Scan(&tableName, &columnName, &pgColumn.ColumnType, &pgColumn.DataTypeOID)
		if err != nil {
			return pgTables, fmt.Errorf("can't map row to values, schema=%s, error=%w", schemaName, err)
		}
		pgTable, ok := pgTables[tableName]
		if !ok {
			pgTable = make(PGTable)
		}
		pgTable[columnName] = pgColumn
		pgTables[tableName] = pgTable
	}
	if len(pgTables) == 0 {
		return pgTables, fmt.Errorf("cannot read destination metadata, check user rights, schema=%s", schemaName)
	}
	log.Debug("Got tables", "tables", pgTables)

	// Assign primary keys
	// for tableName, pgTable := range pgTables {
	// 	var pk map[string]bool
	// 	pk, err = getPrimaryKey(log, database, schemaName, tableName)
	// 	if err != nil {
	// 		return pgTables, err
	// 	}
	// 	for columnName, column := range pgTable {
	// 		_, ok := pk[columnName]
	// 		if ok {
	// 			column.PrimaryKey = true
	// 			pgTable[columnName] = column
	// 		}
	// 	}
	// 	pgTables[tableName] = pgTable
	// }
	// log.Debug("Assigned PK", "tables", pgTables)
	return pgTables, nil
}
