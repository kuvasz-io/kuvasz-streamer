package main

import (
	"net/http"
)

type db struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func dbGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item db
	req := PrepareReq(w, r)

	id, err := ExtractId(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	err = ConfigDB.QueryRow(`SELECT db_id, name FROM db WHERE db_id = ?`, id).Scan(&item.Id, &item.Name)
	if err != nil {
		log.Error("Cannot read database schema", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read database schema list", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func dbGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var dbs []db

	req := PrepareReq(w, r)

	rows, err := ConfigDB.Query(`SELECT db_id, name FROM db`)
	if err != nil {
		log.Error("Cannot read database schema list", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read database schema list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item db
		err := rows.Scan(&item.Id, &item.Name)
		if err != nil {
			log.Error("Cannot scan item", "error", err)
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		dbs = append(dbs, item)
	}
	req.ReturnOK(w, r, dbs, len(dbs))
}
