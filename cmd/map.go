package main

import (
	"fmt"
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
		Type   string `yaml:"type,omitempty"`
		Target string `yaml:"target,omitempty"`
	}
)

func MapSourceTable(relationName string, sourceTables map[string]SourceTable) (string, error) {
	var destTable string
	sourceTable, ok := sourceTables[relationName]
	if !ok {
		return "", fmt.Errorf("unconfigured source table=%s", relationName)
	}
	if sourceTable.Target != "" {
		destTable = sourceTable.Target
	} else {
		destTable = relationName
	}
	_, ok = destTables[destTable]
	if !ok {
		return "", fmt.Errorf("destination table does not exist, table=%s", destTable)
	}
	return destTable, nil
}
