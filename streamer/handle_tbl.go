package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type tbl struct {
	ID              int64   `json:"id"`
	DBId            int64   `json:"db_id"`
	DBName          string  `json:"db_name"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Target          string  `json:"target"`
	PartitionsRegex *string `json:"partitions_regex"`
}

var tblColumns = map[string]string{
	"id":               "tbl.tbl_id",
	"db_id":            "tbl.db_id",
	"name":             "tbl.name",
	"type":             "tbl.type",
	"target":           "tbl.target",
	"partitions_regex": "tbl.partitions_regex",
}

func tblGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item tbl
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	err = ConfigDB.QueryRow(
		`SELECT tbl.tbl_id, tbl.db_id, db.name as db_name, tbl.name, tbl.type, tbl.target, tbl.partitions_regex 
		FROM tbl INNER JOIN DB on tbl.db_id = db.db_id
		WHERE tbl_id = ?`,
		id).Scan(&item.ID, &item.DBId, &item.DBName, &item.Name, &item.Type, &item.Target, &item.PartitionsRegex)
	if errors.Is(err, sql.ErrNoRows) {
		req.ReturnError(w, http.StatusNotFound, "not_found", "can't find table", nil)
		return
	}
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func tblGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var tbls []tbl

	req := PrepareReq(w, r)

	m := ValuesToModifier(r.URL.Query(), tblColumns)
	query := BuildQuery(
		`SELECT tbl.tbl_id, tbl.db_id, db.name as db_name, tbl.name, tbl.type, tbl.target, tbl.partitions_regex 
		FROM tbl INNER JOIN DB on tbl.db_id = db.db_id`,
		m)
	log.Debug("running query", "query", query, "modifier", m, "values", r.URL.Query())
	rows, err := ConfigDB.Query(query)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item tbl
		err := rows.Scan(&item.ID, &item.DBId, &item.DBName, &item.Name, &item.Type, &item.Target, &item.PartitionsRegex)
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		tbls = append(tbls, item)
	}
	if err = rows.Err(); err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan tbl item", err)
	}

	req.ReturnOK(w, r, tbls, len(tbls))
}

func tblPostOneHandler(w http.ResponseWriter, r *http.Request) {
	var item tbl
	req := PrepareReq(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body: %v", err)
		req.ReturnError(w, http.StatusInternalServerError, "0000", "Cannot read request", err)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("could not decode tbl", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "0003", "JSON parse error", err)
		return
	}
	if item.DBId == 0 || item.Name == "" || item.Type == "" || item.Target == "" {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Missing parameters", nil)
		return
	}
	log.Debug("Creating tbl", "item", item)
	// err = app.Validate.Struct(item)

	result, err := ConfigDB.Exec(
		`INSERT INTO tbl(db_id, name, type, target, partitions_regex) VALUES (?, ?, ?, ?, ?)`,
		item.DBId, item.Name, item.Type, item.Target, item.PartitionsRegex)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	item.ID, _ = result.LastInsertId()
	log.Debug("Created tbl", "item", item)
	req.ReturnOK(w, r, item, 1)
}

func tblDeleteOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	result, err := ConfigDB.Exec(`DELETE FROM tbl WHERE tbl_id = ?`, id)
	if err != nil {
		log.Error("Cannot delete tbl", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't delete tbl", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "url not found", nil)
		return
	}
	req.ReturnOK(w, r, nil, 0)
}

func tblPutOneHandler(w http.ResponseWriter, r *http.Request) {
	var item tbl
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body: %v", err)
		req.ReturnError(w, http.StatusInternalServerError, "0000", "Cannot read request", err)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("could not decode tbl", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "0003", "JSON parse error", err)
		return
	}

	// err = app.Validate.Struct(item)
	if item.DBId == 0 || item.Name == "" || item.Type == "" || item.Target == "" {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Missing parameters", nil)
		return
	}

	log.Debug("Updating tbl", "id", id, "item", item)

	result, err := ConfigDB.Exec(
		`UPDATE tbl set name=?, type=?, target=?, partitions_regex=? where tbl_id=?`,
		item.Name, item.Type, item.Target, item.PartitionsRegex, id)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "tbl not found", nil)
		return
	}

	req.ReturnOK(w, r, item, 1)
}
