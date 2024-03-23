package main

import (
	"context"
	"fmt"
	"net/http"
)

func MapGetOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id
	id, err := ExtractID(r)

	// Validate id
	if len(MappingTable) == 0 {
		err := RefreshMappingTable()
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
			return
		}
	}
	if err != nil || id < 0 || id >= int64(len(MappingTable)) {
		req.ReturnError(w, http.StatusNotFound, "invalid_id", "Invalid ID", err)
		return
	}
	log := log.With("id", id)
	result := FindTableByID(id)
	if result.DBName == "" {
		req.ReturnError(w, http.StatusNotFound, "not_found", "Not found", nil)
		return
	}
	log.Debug("Got one map", "result", result)
	req.ReturnOK(w, r, result, 1)
}

func MapGetManyHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	if len(MappingTable) == 0 {
		err := RefreshMappingTable()
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
			return
		}
	}
	req.ReturnOK(w, r, MappingTable, len(MappingTable))
}

func MapCreateTableHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id
	id, err := ExtractID(r)
	if err != nil || id < 0 || id >= int64(len(MappingTable)) {
		req.ReturnError(w, http.StatusNotFound, "invalid_id", "Invalid ID", err)
		return
	}
	t := MappingTable[id]
	if t.Present {
		req.ReturnError(
			w,
			http.StatusBadRequest,
			"conflict",
			"destination table already present",
			fmt.Errorf("destination table %s already present", t.Name))
		return
	}
	q := "CREATE TABLE " + t.Name + "(sid text"
	for k, v := range t.SourceColumns {
		q += ", " + k + " " + v.ColumnType
	}
	q += ");"
	_, err = DestConnectionPool.Exec(context.Background(), q)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "cannot create table", q, err)
		return
	}
	err = RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}
	req.ReturnOK(w, r, t, 1)
}

func MapCloneTableHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id
	id, err := ExtractID(r)
	if err != nil || id < 0 || id >= int64(len(MappingTable)) {
		req.ReturnError(w, http.StatusNotFound, "invalid_id", "Invalid ID", err)
		return
	}
	t := MappingTable[id]
	log = log.With("handler", "MapCloneTableHandler", "id", id)

	log.Debug("Cloning table", "name", t.Name)
	// Create table if not present
	if !t.Present {
		q := "CREATE TABLE " + t.Name + "(sid text"
		for k, v := range t.SourceColumns {
			q += ", " + k + " " + v.ColumnType
		}
		q += ");"
		log.Debug("Creating table", "name", t.Name, "columns", t.SourceColumns, "q", q)
		_, err = DestConnectionPool.Exec(context.Background(), q)
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "cannot create table", q, err)
			return
		}
	}
	// Now add it to config
	log.Debug("Adding entry to tbl", "db_id", t.DBId, "name", t.Name)
	_, err = ConfigDB.Exec(
		`INSERT INTO tbl(db_id, name, type, target) VALUES (?, ?, ?, ?)`,
		t.DBId, t.Name, "clone", t.Name)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "cannot add entry to tbl", err)
		return
	}

	// Finally refresh mapping table
	err = RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}
	req.ReturnOK(w, r, t, 1)
}

func MapRefreshHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	err := RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}
	req.ReturnOK(w, r, nil, 0)
}
