package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"sort"

	"github.com/jackc/pgx/v5"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

type (
	DBMap          []SourceDatabase
	SourceDatabase struct {
		ID     int64                  `json:"db_id"`
		Name   string                 `yaml:"database" json:"database"`
		Urls   []SourceURL            `yaml:"urls"     json:"urls"`
		Tables map[string]SourceTable `yaml:"tables"   json:"tables"`
	}
	SourceURL struct {
		ID      int64  `json:"url_id"`
		URL     string `yaml:"url"     json:"url"`
		SID     string `yaml:"sid"     json:"sid"`
		Version int    `json:"version"`
	}
	SourceTable struct {
		ID              int64  `json:"tbl_id"`
		Type            string `yaml:"type,omitempty"             json:"type"`
		Target          string `yaml:"target,omitempty"           json:"target"`
		PartitionsRegex string `yaml:"partitions_regex,omitempty" json:"partitions_regex"`
		compiledRegex   *regexp.Regexp
	}
)

type mappingEntry struct {
	ID              int64               `json:"id"`
	DBId            int64               `json:"db_id"`
	DBName          string              `json:"db_name"`
	Name            string              `json:"name"`
	Type            string              `json:"type"`
	Target          string              `json:"target"`
	Partitions      []string            `json:"partitions"`
	PartitionsRegex *string             `json:"partitions_regex"`
	Replicated      bool                `json:"replicated"`
	Present         bool                `json:"present"`
	SourceColumns   map[string]PGColumn `json:"source_columns"`
	DestColumns     map[string]PGColumn `json:"dest_columns"`
}

type mappingTable []mappingEntry

func (m mappingTable) Len() int { return len(m) }
func (m mappingTable) Less(i, j int) bool {
	if m[i].DBId < m[j].DBId {
		return true
	}
	if m[i].DBId > m[j].DBId {
		return false
	}
	return m[i].Name < m[j].Name
}
func (m mappingTable) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

var MappingTable mappingTable

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
				t.name,
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
	log.Debug("Assigning IDs for yaml map file")
	var dbid, urlid, tblid int64
	dbid = 1
	urlid = 1
	tblid = 1
	for k, db := range dbmap {
		db.ID = dbid
		dbmap[k] = db
		dbid++
		for k, url := range db.Urls {
			url.ID = urlid
			db.Urls[k] = url
			urlid++
		}
		for k, v := range db.Tables {
			v.ID = tblid
			if v.Type == "" {
				v.Type = "clone"
			}
			if v.Target == "" {
				v.Target = k
			}
			db.Tables[k] = v
			tblid++
		}
	}
	log.Info("Fixed map file", "map", dbmap)
}

func CompileRegexes() {
	log.Debug("Compiling partition regexes")
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

func findConfiguredTable(m DBMap, dbID int64, name string) SourceTable {
	var t = SourceTable{}
	for _, db := range m {
		if db.ID != dbID {
			continue
		}
		t = db.Tables[name]
		return t
	}
	return t
}

func getSourceTables(log *slog.Logger, s SourceDatabase) (PGTables, error) {
	if len(s.Urls) == 0 {
		return PGTables{}, nil
	}
	u := s.Urls[0]
	log = log.With("url", u)
	parsedConfig, err := pgx.ParseConfig(u.URL)
	if err != nil {
		return PGTables{}, fmt.Errorf("error parsing database url=%s, error=%s", u.URL, err)
	}
	parsedConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(context.Background(), parsedConfig)
	if err != nil {
		return PGTables{}, fmt.Errorf("error connecting to database=%s, error=%s", u.URL, err)
	}
	defer conn.Close(context.Background())
	sourceTables, err := GetTables(log, conn, "public")
	if err != nil {
		return PGTables{}, fmt.Errorf("error getting tables, error=%s", err)
	}
	return sourceTables, nil
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

func RefreshMappingTable() error {
	// Step 1. Get list of destination tables
	destConn, err := DestConnectionPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("can't acquire connection to destination database, error=%w", err)
	}
	defer destConn.Release()

	destinationTables, err := GetTables(log, destConn.Conn(), "public")
	if err != nil {
		return fmt.Errorf("can't get destination table metadata, error=%w", err)
	}

	// Step 2. Get configured database map
	configuredMap, err := ReadMapDatabase(ConfigDB)
	if err != nil {
		return fmt.Errorf("can't read database map, error=%w", err)
	}

	// Step 3. Loop over provided URLs, get source tables and merge
	var result mappingTable

	for _, db := range configuredMap {
		sourceTables, err := getSourceTables(log, db)
		if err != nil {
			log.Error("Can't get source table metadata", "error", err)
			continue
		}
		if len(sourceTables) == 0 {
			log.Debug("No source tables found", "db", db.Name)
			continue
		}
		for k := range sourceTables {
			configuredTable := findConfiguredTable(configuredMap, db.ID, k)
			t := mappingEntry{
				DBId:            db.ID,
				DBName:          db.Name,
				Name:            k,
				Type:            configuredTable.Type,
				Target:          configuredTable.Target,
				Partitions:      sourceTables[k].Partitions,
				PartitionsRegex: &configuredTable.PartitionsRegex,
				Replicated:      (configuredTable.Type != ""),
				SourceColumns:   sourceTables[k].Columns,
			}
			destName := configuredTable.Target
			if destName == "" {
				destName = k
			}
			d, ok := destinationTables[destName]
			if t.Type != "" || ok {
				t.Present = true
				t.DestColumns = d.Columns
			}
			result = append(result, t)
		}
	}
	sort.Sort(result)
	for i := range result {
		result[i].ID = int64(i)
	}
	MappingTable = result
	log.Debug("Refreshed mapping table", "count", len(MappingTable))
	return nil
}

func FindTableByID(id int64) mappingEntry {
	for i := range MappingTable {
		if MappingTable[i].ID == id {
			return MappingTable[i]
		}
	}
	return mappingEntry{}
}
