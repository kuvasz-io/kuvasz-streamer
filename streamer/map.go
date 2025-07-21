package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type (
	DBMap          []SourceDatabase
	SourceDatabase struct {
		ID     int64        `json:"db_id"`
		Name   string       `json:"database" yaml:"database"`
		Urls   []SourceURL  `json:"urls"     yaml:"urls"`
		Tables SourceTables `json:"tables"   yaml:"tables"`
	}
	SourceURL struct {
		ID             int64  `json:"url_id"`
		URL            string `json:"url"     yaml:"url"`
		SID            string `json:"sid"     yaml:"sid"`
		Version        int    `json:"version"`
		commandChannel chan string
	}
	SourceTable struct {
		ID              int64             `json:"tbl_id"`
		Type            string            `json:"type"             yaml:"type,omitempty"`
		Target          string            `json:"target"           yaml:"target,omitempty"`
		Filter          string            `json:"filter"           yaml:"filter,omitempty"`
		Set             map[string]string `json:"set"              yaml:"set,omitempty"`
		Insert          string            `json:"insert"           yaml:"insert,omitempty"`
		PartitionsRegex string            `json:"partitions_regex" yaml:"partitions_regex,omitempty"`
		compiledRegex   *regexp.Regexp
	}
	SourceTables map[string]SourceTable
)

func ReadMapDatabase(db *sql.DB) (DBMap, error) {
	var jsonData string
	fullMap := DBMap{}
	log := log.With("database", config.App.MapDatabase)
	log.Info("Reading map database")
	err := db.QueryRow(`SELECT json_group_array(
		json_object(
		  'db_id', d.db_id,
		  'database', d.name,
		  'urls', (
			SELECT json_group_array(
			  json_object(
				'url_id', u.url_id,
				'url', u.url,
				'sid', u.sid
			  )
			)
			FROM url u
			WHERE u.db_id = d.db_id
		  ),
		  'tables', (
			SELECT json_group_object(
				t.schema || "." || t.name,
			  json_object(
				'tbl_id', t.tbl_id,
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
		log.Error("Can't read config database", "error", err)
		return fullMap, fmt.Errorf("can't read database, error=%w", err)
	}
	err = json.Unmarshal([]byte(jsonData), &fullMap)
	if err != nil {
		return fullMap, fmt.Errorf("can't unmarshal config database, error=%w", err)
	}
	log.Info("Read map database", "map", fullMap)
	return fullMap, nil
}

func ReadMapFile(filename string) (DBMap, error) {
	var err error
	var data []byte
	var m DBMap

	// Read map
	log := log.With("filename", filename)
	log.Info("Reading map file")
	data, err = os.ReadFile(filename)
	if err != nil {
		return m, fmt.Errorf("can't read map file, error=%w", err)
	}
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return m, fmt.Errorf("can't unmarshal map file, error=%w", err)
	}
	log.Info("Read map file", "map", m)
	log.Debug("Assigning IDs for yaml map file")
	var dbid, urlid, tblid int64
	dbid = 1
	urlid = 1
	tblid = 1
	for k, db := range m {
		db.ID = dbid
		dbid++
		for k, url := range db.Urls {
			url.ID = urlid
			db.Urls[k] = url
			urlid++
		}
		t := make(map[string]SourceTable)
		for k, v := range db.Tables {
			schema, table := splitSchema(k)
			v.ID = tblid
			if v.Type == "" {
				v.Type = "clone"
			}
			if v.Target == "" {
				v.Target = table
			}
			t[joinSchema(schema, table)] = v
			tblid++
		}
		db.Tables = t
		m[k] = db
	}
	log.Info("Fixed map file", "map", m)
	return m, nil
}

func (m DBMap) CompileRegexes() {
	log.Debug("Compiling partition regexes")
	for _, db := range m {
		for k, v := range db.Tables {
			if v.PartitionsRegex != "" {
				re, err := regexp.Compile(v.PartitionsRegex)
				if err != nil {
					log.Error("Invalid partition regex", "table", k, "regex", v.PartitionsRegex)
					os.Exit(1)
				}
				v.compiledRegex = re
			}
			db.Tables[k] = v
		}
	}
}

func ReadMap() {
	var err error
	if config.App.MapDatabase != "" {
		SetupConfigDB()
		Migrate(embedMigrations, "migrations", ConfigDB)
		dbmap, err = ReadMapDatabase(ConfigDB)
		if err != nil {
			log.Error("Can't read config database", "error", err)
			os.Exit(1)
		}
		err = RefreshMappingTable()
		if err != nil {
			log.Error("Can't refresh mapping table", "error", err)
			os.Exit(1)
		}
	} else {
		dbmap, err = ReadMapFile(config.App.MapFile)
		if err != nil {
			log.Error("Can't read map file", "error", err)
			os.Exit(1)
		}
		err = RefreshMappingTable()
		if err != nil {
			log.Error("Can't refresh mapping table", "error", err)
			os.Exit(1)
		}
	}
}

func (s SourceTables) Find(t string) string {
	// Quick path for exact match
	_, ok := s[t]
	if ok {
		return t
	}
	// Now try regex
	for sourceTableName, sourceTable := range s {
		if sourceTable.compiledRegex == nil {
			continue
		}
		if sourceTable.compiledRegex.MatchString(t) {
			return sourceTableName
		}
	}
	return ""
}

func (s SourceTables) GetTable(table string) (string, error) {
	var destTable string
	log.Debug("GetTable", "table", table, "sourceTables", s, "DestTables", DestTables)
	sourceTable := s.Find(table)
	if sourceTable == "" {
		return "", fmt.Errorf("unconfigured source table=%s", table)
	}
	t := s[sourceTable]
	if t.Target == "" {
		destTable = sourceTable
	} else {
		destTable = joinSchema(config.Database.Schema, t.Target)
	}
	_, ok := DestTables[destTable]
	if !ok {
		return "", fmt.Errorf("destination table does not exist, table=%s", destTable)
	}
	return destTable, nil
}

func (m DBMap) findConfiguredTable(dbID int64, name string) SourceTable {
	for _, db := range m {
		if db.ID != dbID {
			continue
		}
		return db.Tables[name]
	}
	return SourceTable{}
}

func getSourceTables(log *slog.Logger, s SourceDatabase) (PGTables, error) {
	if len(s.Urls) == 0 {
		return PGTables{}, nil
	}
	u := s.Urls[0]
	log = log.With("url", u)
	parsedConfig, err := pgx.ParseConfig(u.URL)
	if err != nil {
		return PGTables{}, fmt.Errorf("error parsing database url=%s, error=%w", u.URL, err)
	}
	parsedConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(context.Background(), parsedConfig)
	if err != nil {
		return PGTables{}, fmt.Errorf("error connecting to database=%s, error=%w", u.URL, err)
	}
	defer conn.Close(context.Background())
	sourceTables, err := GetTables(log, conn, "%")
	if err != nil {
		return PGTables{}, fmt.Errorf("error getting tables, error=%w", err)
	}
	return sourceTables, nil
}
