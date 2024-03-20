package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

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

func MapGetOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	// extract id
	id, err := ExtractID(r)
	if err != nil || id < 0 || id >= int64(len(MappingTable)) {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}
	log := log.With("id", id)
	result := FindTableByID(id)
	if result.DBName == "" {
		log.Debug("Not found")
		req.ReturnError(w, http.StatusNotFound, "not_found", "Not found", nil)
		return
	}
	log.Debug("Got one map", "result", result)
	req.ReturnOK(w, r, result, 1)

}

func MapGetManyHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	err := RefreshMappingTable()
	if err != nil {
		log.Error("error refreshing mapping table", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}
	req.ReturnOK(w, r, MappingTable, len(MappingTable))
}
