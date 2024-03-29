package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/jackc/pgx/v5"
)

type Publications []string

func findBaseTables(db string) []string {
	var p []string
	for i := range MappingTable {
		// Check table is being replicated and belongs to us
		if !MappingTable[i].Replicated || (MappingTable[i].DBName != db) {
			continue
		}
		// if this is a partitioned table, add all partitions
		if len(MappingTable[i].Partitions) > 0 {
			p = append(p, MappingTable[i].Partitions...)
		} else {
			p = append(p, MappingTable[i].Name)
		}
	}
	return p
}

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
			p += strings.Join(MappingTable[i].Partitions, ", ") + ", "
		} else {
			p = p + MappingTable[i].Name + ", "
		}
	}
	return p[0 : len(p)-2]
}

func SyncPublications(log *slog.Logger, conn *pgx.Conn, db SourceDatabase, schema string) ([]string, error) {
	var newTables []string
	ctx := context.Background()
	publishedTables := mapset.NewSet[string]()

	log.Debug("SyncPublications, step 1: remove unconfigured tables")
	// Fetch list of published tables
	rows, err := conn.Query(
		ctx,
		"SELECT tablename FROM pg_publication_tables WHERE pubname = 'kuvasz_'||$1 and schemaname = $2",
		db.Name,
		schema)
	if err != nil {
		return newTables, fmt.Errorf("cannot query publication tables, error: %w", err)
	}
	defer rows.Close()
	var table string
	for rows.Next() {
		err := rows.Scan(&table)
		if err != nil {
			return newTables, fmt.Errorf("cannot scan table name, error: %w", err)
		}
		publishedTables.Add(table)
		// remove from publication if not in MappingTable, checking for partitions
		if FindSourceTable(table, db.Tables) == "" {
			log.Debug("Removing table from publication", "database", db.Name, "table", table)
			_, err = conn.Exec(ctx, "ALTER PUBLICATION kuvasz_"+db.Name+" DROP TABLE "+schema+"."+table)
			if err != nil {
				return newTables, fmt.Errorf("cannot alter publication, error: %w", err)
			}
		}
	}
	log.Debug("Published tables", "tables", publishedTables)
	log.Debug("Configured tables", "tables", db.Tables)
	// Now add tables missing from publication
	log.Debug("SyncPublications, step 2: add missing tables")
	p := findBaseTables(db.Name)
	log.Debug("Got base tables, scanning for missing ones", "basetables", p)
	for i := range p {
		if !publishedTables.Contains(p[i]) {
			log.Debug("Adding table to publication", "database", db.Name, "table", p[i])
			_, err = conn.Exec(ctx, "ALTER PUBLICATION kuvasz_"+db.Name+" ADD TABLE "+schema+"."+p[i])
			if err != nil {
				return newTables, fmt.Errorf("cannot alter publication, error: %w", err)
			}
			newTables = append(newTables, p[i])
		}
	}
	return newTables, nil
}
