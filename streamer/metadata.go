package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type (
	PGColumn struct {
		Name        string `json:"name"`
		ColumnType  string `json:"column_type"`
		DataTypeOID uint32 `json:"data_type_oid"`
		PrimaryKey  bool   `json:"primary_key"`
	}
	PGTable struct {
		Columns    map[string]PGColumn
		Partitions []string
	}
	PGTables map[string]PGTable

	PGRelation struct {
		Namespace    string
		RelationName string
		Columns      []PGColumn
	}
	PGRelations map[uint32]PGRelation
)

func getPrimaryKey(log *slog.Logger, database *pgx.Conn, schemaName string, tableName string) (map[string]bool, error) {
	result := make(map[string]bool)
	// query := `SELECT c.column_name FROM information_schema.table_constraints tc
	// 		  JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)
	// 		  JOIN information_schema.columns AS c
	// 		    ON c.table_schema = tc.constraint_schema  AND tc.table_name = c.table_name AND ccu.column_name = c.column_name
	// 		  WHERE constraint_type = 'PRIMARY KEY'
	// 		  	and tc.constraint_catalog =current_database()
	// 			and c.table_schema = $1
	// 			and tc.table_name = $2`
	query := `SELECT a.attname 
	FROM pg_index i 
		JOIN pg_class c ON c.oid = i.indrelid 
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = any(i.indkey) 
		JOIN pg_namespace n ON n.oid = c.relnamespace 
	WHERE c.relname = $1 
		AND n.nspname = $2
		AND i.indisprimary;`
	pkRows, err := database.Query(context.Background(), query, tableName, schemaName)
	if err != nil {
		log.Error("cannot get primary keys", "error", err)
		return result, fmt.Errorf("cannot get primary keys, table=%s, error=%w", tableName, err)
	}
	defer pkRows.Close()

	for pkRows.Next() {
		var columnName string

		err = pkRows.Scan(&columnName)
		if err != nil {
			return result, fmt.Errorf("cannot map row constraints to values, table=%s, column=%s, error=%w", tableName, columnName, err)
		}
		result[columnName] = true
	}
	return result, nil
}

func GetTables(log *slog.Logger, database *pgx.Conn, schemaName string) (PGTables, error) {
	query := `WITH p as (
				SELECT inhparent as table, array_agg (inhrelid::pg_catalog.regclass) as partitions
				FROM pg_catalog.pg_inherits
				GROUP BY 1) 
			  SELECT c.table_name, c.column_name, c.udt_name, t.oid, p.partitions
				  FROM information_schema.columns as c
					  INNER JOIN pg_type as t ON c.udt_name=t.typname
					  INNER JOIN pg_catalog.pg_class as pg ON pg.relname=c.table_name
					  LEFT JOIN p on p.table=pg.oid
				  WHERE c.table_catalog=current_database() 
					and not pg.relispartition
					and pg.relkind in ('r', 'p')
					and c.table_schema=$1;`

	pgTables := make(PGTables)
	if database == nil {
		return pgTables, fmt.Errorf("no connection to database")
	}
	log = log.With("schema", schemaName)
	log.Debug("Fetching tables and columns")
	rows, err := database.Query(context.Background(), query, schemaName)
	if err != nil {
		return pgTables, fmt.Errorf("cannot get column metadata from database, schema=%s, error=%w", schemaName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var pgColumn PGColumn
		var columnName string
		var partitions []string
		err = rows.Scan(&tableName, &columnName, &pgColumn.ColumnType, &pgColumn.DataTypeOID, &partitions)
		if err != nil {
			return pgTables, fmt.Errorf("can't map row to values, schema=%s, error=%w", schemaName, err)
		}
		pgTable, ok := pgTables[tableName]
		if !ok {
			pgTable.Partitions = partitions
			pgTable.Columns = make(map[string]PGColumn)
		}
		pgTable.Columns[columnName] = pgColumn
		pgTables[tableName] = pgTable
	}
	if len(pgTables) == 0 {
		return pgTables, fmt.Errorf("empty destination metadata, check user rights and destination database schema, schema=%s", schemaName)
	}
	log.Debug("Got tables", "tables", pgTables)

	// Assign primary keys
	for tableName, pgTable := range pgTables {
		var pk map[string]bool
		pk, err = getPrimaryKey(log, database, schemaName, tableName)
		if err != nil {
			return pgTables, err
		}
		for columnName, column := range pgTable.Columns {
			_, ok := pk[columnName]
			if ok {
				column.PrimaryKey = true
				pgTable.Columns[columnName] = column
			}
		}
		pgTables[tableName] = pgTable
	}
	log.Debug("Assigned PK", "tables", pgTables)
	return pgTables, nil
}
