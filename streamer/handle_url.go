package main

import (
	"net/http"
)

type url struct {
	Id   int64  `json:"id"`
	DBId int64  `json:"db_id"`
	SID  string `json:"sid"`
	URL  string `json:"url"`
}

func urlGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item url
	req := PrepareReq(w, r)

	id, err := ExtractId(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	err = ConfigDB.QueryRow(
		`SELECT url_id, db_id, sid, url FROM url WHERE url_id = ?`,
		id).Scan(&item.Id, &item.DBId, &item.SID, &item.URL)
	if err != nil {
		log.Error("Cannot read url", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func urlGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var urls []url

	req := PrepareReq(w, r)

	rows, err := ConfigDB.Query(`SELECT url_id, db_id, sid, url FROM url`)
	if err != nil {
		log.Error("Cannot read url list", "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read url list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item url
		err := rows.Scan(&item.Id, &item.DBId, &item.SID, &item.URL)
		if err != nil {
			log.Error("Cannot scan item", "error", err)
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		urls = append(urls, item)
	}
	req.ReturnOK(w, r, urls, len(urls))
}
