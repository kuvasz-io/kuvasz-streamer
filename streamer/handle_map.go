package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

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

func TablesGetHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}
	u := findURL(id)
	if u == nil {
		log.Error("url not found")
		req.ReturnError(w, http.StatusNotFound, "url_not_found", "URL not found", fmt.Errorf("url %d not found", id))
		return
	}

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
	tbl, err := GetTables(log, conn, "public")
	if err != nil {
		log.Error("error getting tables")
		req.ReturnError(w, http.StatusInternalServerError, "error_getting_tables", "Error getting tables", err)
		return
	}
	req.ReturnOK(w, r, tbl, 1)
}
