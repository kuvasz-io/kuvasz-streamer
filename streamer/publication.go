package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type Publications []string

func makePublication(database SourceDatabase) string {
	log.Debug("Creating publication", "database", database, "MappingTable", MappingTable)
	if len(database.Tables) == 0 {
		return ""
	}
	p := " for table "
	for i := range MappingTable {
		// Check table is being replicated and belongs to us
		if !MappingTable[i].Replicated || (MappingTable[i].DBName != database.Name) {
			continue
		}
		// if this is a partitioned table, add all partitions
		if len(MappingTable[i].Partitions) > 0 {
			for j := range MappingTable[i].Partitions {
				p = p + MappingTable[i].Partitions[j] + ", "
			}
		} else {
			p = p + MappingTable[i].Name + ", "
		}
	}
	return p[0 : len(p)-2]
}

func SyncPublications(log *slog.Logger, conn *pgx.Conn, db SourceDatabase, schema string) error {
	ctx := context.Background()
	publishedTables := make(map[string]bool)
	// Fetch list of published tables
	rows, err := conn.Query(ctx, "SELECT tablename FROM pg_publication_tables WHERE pubname = 'kuvasz_'||$1 and schemaname = $2", db.Name, schema)
	if err != nil {
		return fmt.Errorf("cannot query publication tables, error: %w", err)
	}
	defer rows.Close()
	var table string
	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			return fmt.Errorf("cannot scan table name, error: %w", err)
		}
		publishedTables[table] = true
		// remove from publication if not in MappingTable
		_, ok := db.Tables[table]
		if !ok {
			log.Debug("Removing table from publication", "database", db.Name, "table", table)
			_, err = conn.Exec(ctx, "ALTER PUBLICATION kuvasz_"+db.Name+" DROP TABLE "+schema+"."+table)
			if err != nil {
				return fmt.Errorf("cannot alter publication, error: %w", err)
			}
		}
	}
	log.Debug("Published tables", "tables", publishedTables)
	log.Debug("Configured tables", "tables", db.Tables)
	// Now add tables missing from publication
	for table := range db.Tables {
		_, ok := publishedTables[table]
		if !ok {
			log.Debug("Adding table to publication", "database", db.Name, "table", table)
			_, err = conn.Exec(ctx, "ALTER PUBLICATION kuvasz_"+db.Name+" ADD TABLE "+schema+"."+table)
			if err != nil {
				return fmt.Errorf("cannot alter publication, error: %w", err)
			}
		}
	}
	return nil
}
