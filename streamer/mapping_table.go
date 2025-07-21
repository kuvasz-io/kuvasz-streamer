package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/google/cel-go/cel"
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
	Filter          string              `json:"filter"`
	Set             map[string]string   `json:"set"`
	Partitions      []string            `json:"partitions"`
	PartitionsRegex *string             `json:"partitions_regex"`
	Replicated      bool                `json:"replicated"`
	Present         bool                `json:"present"`
	SourceColumns   map[string]PGColumn `json:"source_columns"`
	DestColumns     map[string]PGColumn `json:"dest_columns"`
	compiledRegex   *regexp.Regexp
	compiledFilter  cel.Program
	compiledSet     map[string]cel.Program
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

func (m mappingTable) FindByID(id int64) MappingEntry {
	for i := range m {
		if m[i].ID == id {
			return m[i]
		}
	}
	return MappingEntry{}
}

func (e MappingEntry) Match(t string) bool {
	// Quick path for exact match
	if e.Name == t {
		return true
	}
	// Now try regex
	if e.compiledRegex == nil {
		return false
	}
	if e.compiledRegex.MatchString(t) {
		return true
	}
	return false
}

func (m mappingTable) FindByName(db string, name string) (MappingEntry, error) {
	schema, table := splitSchema(name)
	log.Debug("Finding table", "MappingTable", m, "db", db, "name", name, "schema", schema, "table", table)
	for i := range m {
		if m[i].DBName == db && m[i].Schema == schema && m[i].Match(table) {
			return m[i], nil
		}
	}
	return MappingEntry{}, fmt.Errorf("table not found, db:%s, table: %s", db, table)
}

var MappingTable mappingTable

func RefreshMappingTable() error { //nolint:gocognit
	var err error
	// Step 1. Get list of destination tables
	destConn, err := DestConnectionPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("can't acquire connection to destination database, error=%w", err)
	}
	defer destConn.Release()

	DestTables, err = GetTables(log, destConn.Conn(), config.Database.Schema)
	if err != nil {
		return fmt.Errorf("can't get destination table metadata while refreshing, error=%w", err)
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
				Filter:          configuredTable.Filter,
				Set:             configuredTable.Set,
				Partitions:      sourceTables[k].Partitions,
				PartitionsRegex: &configuredTable.PartitionsRegex,
				Replicated:      (configuredTable.Type != ""),
				SourceColumns:   sourceTables[k].Columns,
			}
			if t.PartitionsRegex != nil && *t.PartitionsRegex != "" {
				re, err := regexp.Compile(*t.PartitionsRegex)
				if err != nil {
					return fmt.Errorf("can't compile partition regex, table:%s, regex:%s", k, *t.PartitionsRegex)
				}
				t.compiledRegex = re
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
			env := ConvertPGColumnsToEnv(t.SourceColumns)
			if t.Filter != "" {
				t.compiledFilter, err = prepareExpression(t.Filter, env)
				if err != nil {
					return fmt.Errorf("can't compile filter: %s, error: %w", t.Filter, err)
				}
			}
			t.compiledSet = make(map[string]cel.Program)
			for c, p := range t.Set {
				t.compiledSet[c], err = prepareExpression(p, env)
				if err != nil {
					return fmt.Errorf("can't compile Set statement: %s, error: %w", p, err)
				}
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
