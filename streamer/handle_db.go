package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type db struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var dbColumns = map[string]string{
	"id":    "db_id",
	"db_id": "db_id",
	"name":  "name",
}

func dbGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item db
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	err = ConfigDB.QueryRow(`SELECT db_id, name FROM db WHERE db_id = ?`, id).Scan(&item.ID, &item.Name)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read database schema list", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func dbGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var dbs []db

	req := PrepareReq(w, r)

	m := ValuesToModifier(r.URL.Query(), dbColumns)
	query := BuildQuery(`SELECT db_id, name FROM db`, m)
	log.Debug("running query", "query", query, "modifier", m, "values", r.URL.Query())
	rows, err := ConfigDB.Query(query)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read database schema list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item db
		err := rows.Scan(&item.ID, &item.Name)
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		dbs = append(dbs, item)
	}
	if err = rows.Err(); err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan database item", err)
	}
	req.ReturnOK(w, r, dbs, len(dbs))
}

func dbPostOneHandler(w http.ResponseWriter, r *http.Request) {
	var item db
	req := PrepareReq(w, r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body: %v", err)
		req.ReturnError(w, http.StatusInternalServerError, "0000", "Cannot read request", err)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("could not decode db", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "0003", "JSON parse error", err)
		return
	}
	log.Debug("Creating db", "item", item)
	// err = app.Validate.Struct(item)

	result, err := ConfigDB.Exec(
		`INSERT INTO db(name) VALUES (?)`, item.Name)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	item.ID, _ = result.LastInsertId()
	log.Debug("Created db", "item", item)
	req.ReturnOK(w, r, item, 1)
}

func dbDeleteOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	result, err := ConfigDB.Exec(`DELETE FROM db WHERE db_id = ?`, id)
	if err != nil {
		log.Error("Cannot delete database schema", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't delete database schema", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "database schema not found", nil)
		return
	}
	req.ReturnOK(w, r, nil, 0)
}

func dbPutOneHandler(w http.ResponseWriter, r *http.Request) {
	var item db
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
		log.Error("could not decode db", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "0003", "JSON parse error", err)
		return
	}
	log.Debug("Updating db", "item", item)
	// err = app.Validate.Struct(item)

	result, err := ConfigDB.Exec(`UPDATE db set name=? where db_id=?`, item.Name, id)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "database schema not found", nil)
		return
	}

	req.ReturnOK(w, r, item, 1)
}
