package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type (
	DBMap          []SourceDatabase
	SourceDatabase struct {
		Name   string                 `yaml:"database" json:"database"`
		Urls   []SourceURL            `yaml:"urls"     json:"urls"`
		Tables map[string]SourceTable `yaml:"tables"   json:"tables"`
	}
	SourceURL struct {
		URL string `yaml:"url" json:"url"`
		SID string `yaml:"sid" json:"sid"`
	}
	SourceTable struct {
		Type            string `yaml:"type,omitempty"             json:"type"`
		Target          string `yaml:"target,omitempty"           json:"target"`
		PartitionsRegex string `yaml:"partitions_regex,omitempty" json:"partitions_regex"`
		compiledRegex   *regexp.Regexp
		id              int
	}
)

func ReadMapDatabase(db *sql.DB) {
	var jsonData string
	log := log.With("database", config.App.MapFile)
	log.Info("Reading map database")
	err := db.QueryRow(`SELECT json_group_array(
		json_object(
		  'database', d.name,
		  'urls', (
			SELECT json_group_array(
			  json_object(
				'url', u.url,
				'sid', u.sid
			  )
			)
			FROM url u
			WHERE u.db_id = d.db_id
		  ),
		  'tables', (
			SELECT json_group_object(
				t.name,
			  json_object(
				'type', t.type,
				'target', t.target,
				'partitions_regex', t.partitions_regex
			  )
			)
			FROM tbl t
			WHERE t.db_id = d.db_id
		  )
		)
	  )
	  FROM db d;`).Scan(&jsonData)
	if err != nil {
		log.Error("Can't read database", "error", err)
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(jsonData), &dbmap)
	if err != nil {
		log.Error("Can't unmarshal database", "error", err)
		os.Exit(1)
	}
	log.Info("Read map database", "map", dbmap)
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
}

func CompileRegexes() {
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
