package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

func splitSchema(t string) (string, string) {
	before, after, found := strings.Cut(t, ".")
	if found {
		return before, after
	}
	return config.App.DefaultSchema, t
}

func joinSchema(schema, table string) string {
	if schema == "" {
		return table
	}
	return schema + "." + table
}

func getPrimaryKey(log *slog.Logger, database *pgx.Conn, tableName string) (map[string]bool, error) {
	result := make(map[string]bool)
	s, t := splitSchema(tableName)
	query := `SELECT a.attname 
	FROM pg_index i 
		JOIN pg_class c ON c.oid = i.indrelid 
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = any(i.indkey) 
		JOIN pg_namespace n ON n.oid = c.relnamespace 
	WHERE c.relname = $1 
		AND n.nspname = $2
		AND i.indisprimary;`
	pkRows, err := database.Query(context.Background(), query, t, s)
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
			  SELECT c.table_schema, c.table_name, c.column_name, c.udt_name, t.oid, p.partitions
				  FROM information_schema.columns as c
					  INNER JOIN pg_type as t ON c.udt_name=t.typname
					  INNER JOIN pg_catalog.pg_class as pg ON pg.relname=c.table_name
					  LEFT JOIN p on p.table=pg.oid
				  WHERE c.table_catalog=current_database() 
					and not pg.relispartition
					and pg.relkind in ('r', 'p')
					and c.table_schema like $1
					and c.table_schema not like 'pg_%'
					and c.table_schema <> 'information_schema';`

	pgTables := make(PGTables)
	if database == nil {
		return pgTables, errors.New("no connection to database")
	}
	log = log.With("schema", schemaName)
	log.Debug("Fetching tables and columns")
	rows, err := database.Query(context.Background(), query, schemaName)
	if err != nil {
		return pgTables, fmt.Errorf("cannot get column metadata from database, schema=%s, error=%w", schemaName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var s, t string
		var pgColumn PGColumn
		var columnName string
		var partitions []string
		err = rows.Scan(&s, &t, &columnName, &pgColumn.ColumnType, &pgColumn.DataTypeOID, &partitions)
		if err != nil {
			return pgTables, fmt.Errorf("can't map row to values, schema=%s, error=%w", schemaName, err)
		}
		tableName := joinSchema(s, t)
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
		pk, err = getPrimaryKey(log, database, tableName)
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

func randomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] //nolint:gosec // this is just for non collision of files
	}
	return string(b)
}

func SetupDestination() error {
	// Create connection pool
	pgconfig, err := pgxpool.ParseConfig(config.Database.URL)
	if err != nil {
		return fmt.Errorf("failed to parse database url:%s, err:%w", config.Database.URL, err)
	}

	// Create origin and setup hook to setup session if supported and required
	if config.Database.Origin != "" {
		origin := "kuvasz_" + config.Database.Origin + "_" + randomString()
		log.Info("Using Origin", "origin", origin)

		pgconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			var oid pgtype.Uint32
			log.Info("Creating Origin", "origin", "kuvasz_"+origin)
			err = conn.QueryRow(ctx, "select pg_replication_origin_create('"+origin+"')").Scan(&oid)
			if err != nil {
				return fmt.Errorf("cannot create origin: %s, error: %w", origin, err)
			}
			_, err := conn.Exec(ctx, "select pg_replication_origin_session_setup('"+origin+"')")
			if err != nil {
				return fmt.Errorf("cannot create origin: %s, error: %w", origin, err)
			}
			return nil
		}
		pgconfig.BeforeClose = func(conn *pgx.Conn) {
			log.Info("Dropping Origin", "origin", "kuvasz_"+origin)
			_, err := conn.Exec(context.Background(), "select pg_replication_origin_session_reset()")
			if err != nil {
				log.Error("cannot reset session", "error", err)
			}
			_, err = conn.Exec(context.Background(), "select pg_replication_origin_session_drop('"+origin+"')")
			if err != nil {
				log.Error("cannot drop origin", "origin", origin, "error", err)
			}
		}
	}
	DestConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgconfig)
	if err != nil {
		return fmt.Errorf("can't connect to target database, url=%s, error=%w", config.Database.URL, err)
	}
	log.Info("Connected to target database", "url", config.Database.URL)

	// Get connection
	conn, err := DestConnectionPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("can't acquire connection to fetch metadata: error=%w", err)
	}
	defer conn.Release()

	// Get destination metadata
	log.Info("Getting destination table metadata")
	DestTables, err = GetTables(log, conn.Conn(), config.Database.Schema)
	if err != nil {
		return fmt.Errorf("can't get destination table metadata during initial setup, error=%w", err)
	}
	return nil
}

func CloseDestination() {
	DestConnectionPool.Close()
}
