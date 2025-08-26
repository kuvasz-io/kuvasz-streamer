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
			for j := range MappingTable[i].Partitions {
				p = append(p, joinSchema(MappingTable[i].Schema, MappingTable[i].Partitions[j]))
			}
		} else {
			p = append(p, joinSchema(MappingTable[i].Schema, MappingTable[i].Name))
		}
	}
	return p
}

func makePublication(database SourceDatabase) string {
	log.Debug("Creating publication - find tables", "database", database, "MappingTable", MappingTable)
	if len(database.Tables) == 0 {
		return ""
	}
	p := findBaseTables(database.Name)
	if len(p) == 0 {
		return ""
	}
	return " for table " + strings.Join(p, ", ")
}

func SyncPublications(log *slog.Logger, conn *pgx.Conn, db SourceDatabase, sid string) ([]string, error) {
	var newTables []string
	ctx := context.Background()
	publishedTables := mapset.NewSet[string]()

	pubName := "kuvasz_" + db.Name + "_" + sid
	pubName = strings.ReplaceAll(pubName, "-", "_")

	log.Debug("SyncPublications", "db", db, "publication", pubName)
	log.Debug("SyncPublications, step 1: Find published tables")
	// Fetch list of published tables
	rows, err := conn.Query(
		ctx,
		"SELECT schemaname,tablename FROM pg_publication_tables WHERE pubname = $1",
		pubName)
	if err != nil {
		return newTables, fmt.Errorf("cannot query publication tables, pubname=%s, error: %w", pubName, err)
	}
	defer rows.Close()
	var schema, table string
	for rows.Next() {
		err := rows.Scan(&schema, &table)
		if err != nil {
			return newTables, fmt.Errorf("cannot scan table name, error: %w", err)
		}
		fullName := joinSchema(schema, table)
		log.Debug("Found published table", "pubName", pubName, "database", db.Name, "schema", schema, "table", table)
		publishedTables.Add(fullName)
		// remove from publication if not in MappingTable, checking for partitions
	}
	log.Debug("Published tables", "tables", publishedTables)
	log.Debug("Configured tables", "tables", db.Tables)
	log.Debug("SyncPublications, step 2: remove unconfigured tables")
	for _, n := range publishedTables.ToSlice() {
		if db.Tables.Find(n) == "" {
			log.Debug("Removing table from publication", "pubName", pubName, "database", db.Name, "schema", schema, "table", table)
			_, err = conn.Exec(ctx, "ALTER PUBLICATION "+pubName+" DROP TABLE "+n)
			if err != nil {
				return newTables, fmt.Errorf("cannot alter publication, pubName=%s, error: %w", pubName, err)
			}
		}
	}

	// Now add tables missing from publication
	log.Debug("SyncPublications, step 3: add missing tables")
	p := findBaseTables(db.Name)
	log.Debug("Got base tables, scanning for missing ones", "basetables", p)
	for i := range p {
		if !publishedTables.Contains(p[i]) {
			log.Debug("Adding table to publication", "pubName", pubName, "database", db.Name, "table", p[i])
			_, err = conn.Exec(ctx, "ALTER PUBLICATION "+pubName+" ADD TABLE "+p[i])
			if err != nil {
				return newTables, fmt.Errorf("cannot alter publication, error: %w", err)
			}
			newTables = append(newTables, p[i])
		}
	}
	return newTables, nil
}
