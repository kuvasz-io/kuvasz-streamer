package main

import (
	"fmt"
	"regexp"
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
		CompiledRegex   *regexp.Regexp
	}
)

func FindSourceTable(relationName string, sourceTables map[string]SourceTable) string {
	// Quick path for exact match
	_, ok := sourceTables[relationName]
	if ok {
		return relationName
	}
	// Now try regex
	for sourceTableName, sourceTable := range sourceTables {
		if sourceTable.CompiledRegex == nil {
			continue
		}
		if sourceTable.CompiledRegex.MatchString(relationName) {
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
