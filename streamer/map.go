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
		ID              int64  `json:"tbl_id"`
		Type            string `json:"type"             yaml:"type,omitempty"`
		Target          string `json:"target"           yaml:"target,omitempty"`
		PartitionsRegex string `json:"partitions_regex" yaml:"partitions_regex,omitempty"`
		compiledRegex   *regexp.Regexp
	}
	SourceTables map[string]SourceTable
)

type MappingEntry struct {
	ID              int64               `json:"id"`
	DBId            int64               `json:"db_id"`
	DBName          string              `json:"db_name"`
	Schema          string              `json:"schema"`
	Table           string              `json:"table"`
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

type mappingTable []MappingEntry

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
			log.Error("Can't read config database, error=%w", err)
			os.Exit(1)
		}
		err = RefreshMappingTable()
		if err != nil {
			log.Error("Can't refresh mapping table, error=%w", err)
			os.Exit(1)
		}
	} else {
		dbmap, err = ReadMapFile(config.App.MapFile)
		if err != nil {
			log.Error("Can't read map file, error=%w", err)
			os.Exit(1)
		}
		err = RefreshMappingTable()
		if err != nil {
			log.Error("Can't refresh mapping table, error=%w", err)
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

func (s SourceTables) GetTable(table string) (*SourceTable, string, error) {
	var destTable string
	log.Debug("GetTable", "table", table, "sourceTables", s, "DestTables", DestTables)
	sourceTable := s.Find(table)
	if sourceTable == "" {
		return nil, "", fmt.Errorf("unconfigured source table=%s", table)
	}
	t := s[sourceTable]
	if t.Target == "" {
		destTable = sourceTable
	} else {
		destTable = joinSchema(config.Database.Schema, t.Target)
	}
	_, ok := DestTables[destTable]
	if !ok {
		return nil, "", fmt.Errorf("destination table does not exist, table=%s", destTable)
	}
	return &t, destTable, nil
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

func RefreshMappingTable() error {
	var err error
	// Step 1. Get list of destination tables
	destConn, err := DestConnectionPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("can't acquire connection to destination database, error=%w", err)
	}
	defer destConn.Release()

	DestTables, err = GetTables(log, destConn.Conn(), config.Database.Schema)
	if err != nil {
		return fmt.Errorf("can't get destination table metadata, error=%w", err)
	}

	// Step 2. Get configured database map
	var configuredMap DBMap
	if config.App.MapDatabase != "" {
		configuredMap, err = ReadMapDatabase(ConfigDB)
		if err != nil {
			return fmt.Errorf("can't read database map, error=%w", err)
		}
	} else {
		configuredMap = dbmap
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
			configuredTable := configuredMap.findConfiguredTable(db.ID, k)
			schema, table := splitSchema(k)
			t := MappingEntry{
				DBId:            db.ID,
				DBName:          db.Name,
				Schema:          schema,
				Name:            table,
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
			d, ok := DestTables[destName]
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
	log.Debug("Refreshed mapping table", "MappingTable", MappingTable)
	return nil
}

func FindTableByID(id int64) MappingEntry {
	for i := range MappingTable {
		if MappingTable[i].ID == id {
			return MappingTable[i]
		}
	}
	return MappingEntry{}
}
