package main

import (
	"net/http"
)

type tbl struct {
	Id              int64 `json:"id"`
	DBId            int64
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Target          string  `json:"target"`
	PartitionsRegex *string `json:"partitions_regex"`
}

func tblGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item tbl
	req := PrepareReq(w, r)

	id, err := ExtractId(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	err = ConfigDB.QueryRow(
		`SELECT tbl_id, db_id, name, type, target, partitions_regex FROM tbl WHERE tbl_id = ?`,
		id).Scan(&item.Id, &item.DBId, &item.Name, &item.Type, &item.Target, &item.PartitionsRegex)
	if err != nil {
		log.Error("Cannot read tbl", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func tblGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var tbls []tbl

	req := PrepareReq(w, r)

	rows, err := ConfigDB.Query(`SELECT tbl_id, db_id, name, type, target, partitions_regex FROM tbl`)
	if err != nil {
		log.Error("Cannot read database schema list", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item tbl
		err := rows.Scan(&item.Id, &item.DBId, &item.Name, &item.Type, &item.Target, &item.PartitionsRegex)
		if err != nil {
			log.Error("Cannot scan item", "error", err)
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		tbls = append(tbls, item)
	}
	req.ReturnOK(w, r, tbls, len(tbls))
}
