package main

import (
	"context"
	"fmt"
	"net/http"
)

func mapGetOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id
	id, err := ExtractID(r)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	// Find id
	if len(MappingTable) == 0 {
		err := RefreshMappingTable()
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
			return
		}
	}
	if id < 0 || id >= int64(len(MappingTable)) {
		req.ReturnError(w, http.StatusNotFound, "invalid_id", "Invalid ID", err)
		return
	}
	log := log.With("id", id)
	result := MappingTable.FindByID(id)
	if result.DBName == "" {
		req.ReturnError(w, http.StatusNotFound, "not_found", "Not found", nil)
		return
	}
	log.Debug("Got one map", "result", result)
	req.ReturnOK(w, r, result, 1)
}

func mapGetManyHandler(w http.ResponseWriter, r *http.Request) {
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

func mapCreateTableHandler(w http.ResponseWriter, r *http.Request) {
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

func mapCloneTableHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	// extract id
	id, err := ExtractID(r)
	if err != nil || id < 0 || id >= int64(len(MappingTable)) {
		req.ReturnError(w, http.StatusNotFound, "invalid_id", "Invalid ID", err)
		return
	}
	t := MappingTable[id]
	log := log.With("handler", "MapReplicateTableHandler", "id", id)

	// extract type
	tabletype := r.URL.Query().Get("type")
	if tabletype == "" {
		tabletype = TableTypeClone
	}
	if tabletype != TableTypeClone && tabletype != TableTypeAppend && tabletype != TableTypeHistory {
		req.ReturnError(w, http.StatusBadRequest, "invalid_type", "Invalid type", nil)
		return
	}

	// extract target name
	target := r.URL.Query().Get("target")
	if target == "" {
		target = t.Name
	}
	fullTargetName := joinSchema(config.Database.Schema, target)

	// extract regex
	regex := r.URL.Query().Get("partitions_regex")
	log.Debug("params", "id", id, "type", tabletype, "target", target, "regex", regex)

	// check the table is not partitioned
	if regex == "" && len(t.Partitions) > 0 {
		req.ReturnError(w, http.StatusBadRequest, "cannot_clone_partitioned_table", "Missing partitions regex", nil)
		return
	}

	log.Debug("Replicating table", "name", t.Name, "type", tabletype, "target", target, "regex", regex, "DestTables", DestTables)

	// Create table if not present
	if _, ok := DestTables[fullTargetName]; !ok {
		q := "CREATE TABLE " + target + "("
		first := true
		for k, v := range t.SourceColumns {
			if first {
				first = false
			} else {
				q += ", "
			}
			q += k + " " + v.ColumnType
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
	log.Debug("Adding entry to tbl", "db_id", t.DBId, "name", t.Name, "target", target, "regex", regex)
	ctx := context.Background()
	_, err = ConfigDB.ExecContext(
		ctx,
		`INSERT INTO tbl(db_id, schema, name, type, target, partitions_regex) VALUES (?, ?, ?, ?, ?, ?)`,
		t.DBId, t.Schema, t.Name, tabletype, target, regex)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "cannot add entry to tbl", err)
		return
	}

	// Now refresh mapping table
	err = RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}

	req.ReturnOK(w, r, t, 1)
}

func mapRefreshHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	err := RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "internal_error", "Error refreshing mapping table", err)
		return
	}
	req.ReturnOK(w, r, nil, 0)
}
