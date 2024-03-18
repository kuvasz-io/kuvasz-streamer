package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type mappingTable struct {
	ID              int64   `json:"id"`
	DBId            int64   `json:"db_id"`
	DBName          string  `json:"db_name"`
	SID             string  `json:"sid"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Target          string  `json:"target"`
	PartitionsRegex *string `json:"partitions_regex"`
	Replicated      bool    `json:"replicated"`
	Present         bool    `json:"present"`
}

func MapGetHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	req.ReturnOK(w, r, dbmap, 1)
}

func findURL(id int64) *SourceURL {
	for _, db := range dbmap {
		for i := range db.Urls {
			if db.Urls[i].ID == id {
				return &db.Urls[i]
			}
		}
	}
	return nil
}

func GetConfiguredTables(log *slog.Logger, urlID int64) ([]mappingTable, error) {
	var tables []mappingTable
	query :=
		`SELECT url.sid, 
				db.db_id,db.name, 
				tbl.tbl_id, tbl.name, tbl.type, tbl.target, tbl.partitions_regex 
		FROM url 
			INNER JOIN db ON url.db_id = db.db_id 
			INNER JOIN tbl ON db.db_id = tbl.db_id
		WHERE url.url_id = ?`
	log.Debug("Getting configured tables")
	rows, err := ConfigDB.Query(query, urlID)
	if err != nil {
		log.Error("Cannot read table list", "query", query, "error", err)
		return tables, err
	}
	defer rows.Close()
	for rows.Next() {
		var item mappingTable
		err := rows.Scan(
			&item.SID,
			&item.DBId, &item.DBName,
			&item.ID, &item.Name, &item.Type, &item.Target, &item.PartitionsRegex)
		if err != nil {
			log.Error("Cannot scan item", "error", err)
			return tables, err
		}
		tables = append(tables, item)
	}
	return tables, nil
}

func findConfiguredTable(configuredTables []mappingTable, k string) mappingTable {
	for _, t := range configuredTables {
		if t.Name == k {
			return t
		}
	}
	return mappingTable{}
}

func TablesGetHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id and get the actual URL
	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}
	log := log.With("url_id", id)
	u := findURL(id)
	if u == nil {
		log.Error("url not found")
		req.ReturnError(w, http.StatusNotFound, "url_not_found", "URL not found", fmt.Errorf("url %d not found", id))
		return
	}

	// Get list of tables from provided URL
	parsedConfig, err := pgx.ParseConfig(u.URL)
	if err != nil {
		log.Error("Error parsing database url", "url", u.URL, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "invalid_config", "Invalid config", err)
		return
	}
	parsedConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(context.Background(), parsedConfig)
	if err != nil {
		log.Error("Cannot start inquiry connection", "databaseURL", u.URL, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "connect_error", "cannot connect", err)
		return
	}
	defer conn.Close(context.Background())
	sourceTables, err := GetTables(log, conn, "public")
	if err != nil {
		log.Error("error getting tables")
		req.ReturnError(w, http.StatusInternalServerError, "error_getting_tables", "Error getting tables", err)
		return
	}

	// Get list of configured tables
	configuredTables, err := GetConfiguredTables(log, id)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "error_getting_config_tables", "Error getting configured tables", err)
	}

	// Now merge the two lists
	var result []mappingTable
	var i int64 = 0
	for k := range sourceTables {
		configuredTable := findConfiguredTable(configuredTables, k)
		t := mappingTable{
			ID:              i,
			DBId:            configuredTable.DBId,
			DBName:          configuredTable.Name,
			SID:             u.SID,
			Name:            k,
			Type:            configuredTable.Type,
			Target:          configuredTable.Target,
			PartitionsRegex: configuredTable.PartitionsRegex,
			Present:         false,
			Replicated:      (configuredTable.Name != ""),
		}
		result = append(result, t)
		i++
	}
	req.ReturnOK(w, r, result, len(result))
}
