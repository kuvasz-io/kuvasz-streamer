package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type (
	DBMap          []SourceDatabase
	SourceDatabase struct {
		Name   string                 `yaml:"database"`
		Urls   []SourceURL            `yaml:"urls"`
		Tables map[string]SourceTable `yaml:"tables"`
	}
	SourceURL struct {
		URL string `yaml:"url"`
		SID string `yaml:"sid"`
	}
	SourceTable struct {
		Type            string `yaml:"type,omitempty"`
		Target          string `yaml:"target,omitempty"`
		PartitionsRegex string `yaml:"partitions_regex,omitempty"`
		compiledRegex   *regexp.Regexp
		id              int
	}
)

func ReadMapDatabase(db *sql.DB) {
	row, err := db.Query("SELECT name FROM db order by db_id;")
	if err != nil {
		log.Error("Can't read database", "error", err)
		os.Exit(1)
	}
	defer row.Close()
	for row.Next() {
		db := SourceDatabase{}
		// Iterate and fetch the records from result cursor
		row.Scan(&db.Name)
	}

}

func ReadMapFile(filename string) {
	// Read map
	log := log.With("filename", filename)
	log.Info("Reading map file")
	var data []byte
	data, err = os.ReadFile(filename)
	if err != nil {
		log.Error("Can't read map file", "error", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(data, &dbmap)
	if err != nil {
		log.Error("Can't unmarshal map file", "error", err)
		os.Exit(1)
	}
	log.Info("Read map file", "map", dbmap)
	log.Debug("Compiling partition regexes and assigning ids")
	i := 0
	for _, db := range dbmap {
		for k, v := range db.Tables {
			if v.PartitionsRegex != "" {
				re, err := regexp.Compile(v.PartitionsRegex)
				if err != nil {
					log.Error("Invalid partition regex", "table", k, "regex", v.PartitionsRegex)
					os.Exit(1)
				}
				v.compiledRegex = re
			}
			v.id = i
			i++
			db.Tables[k] = v
		}
	}
}

func FindSourceTable(relationName string, sourceTables map[string]SourceTable) string {
	// Quick path for exact match
	_, ok := sourceTables[relationName]
	if ok {
		return relationName
	}
	// Now try regex
	for sourceTableName, sourceTable := range sourceTables {
		if sourceTable.compiledRegex == nil {
			continue
		}
		if sourceTable.compiledRegex.MatchString(relationName) {
			return sourceTableName
		}
	}
	return ""
}

func MapSourceTable(relationName string, sourceTables map[string]SourceTable) (*SourceTable, string, error) {
	var destTable string
	sourceTable := FindSourceTable(relationName, sourceTables)
	if sourceTable == "" {
		return nil, "", fmt.Errorf("unconfigured source table=%s", relationName)
	}
	t := sourceTables[sourceTable]
	if t.Target == "" {
		destTable = sourceTable
	} else {
		destTable = t.Target
	}
	_, ok := destTables[destTable]
	if !ok {
		return nil, "", fmt.Errorf("destination table does not exist, table=%s", destTable)
	}
	return &t, destTable, nil
}
