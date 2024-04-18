package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type URL struct {
	ID     int64  `json:"id"`
	DBId   int64  `json:"db_id"`
	DBName string `json:"db_name"`
	SID    string `json:"sid"`
	URL    string `json:"url"`
	Up     bool   `json:"up"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

var URLColumns = map[string]string{
	"id":    "url.db_id",
	"db_id": "url.db_id",
	"sid":   "url.sid",
	"url":   "url.url",
}

func urlGetOneHandler(w http.ResponseWriter, r *http.Request) {
	var item URL
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	// declarative mode
	if config.App.MapDatabase == "" {
		for i := range dbmap {
			for j := range dbmap[i].Urls {
				if dbmap[i].Urls[j].ID == id {
					item.ID = dbmap[i].Urls[j].ID
					item.DBId = dbmap[i].ID
					item.DBName = dbmap[i].Name
					item.SID = dbmap[i].Urls[j].SID
					item.URL = dbmap[i].Urls[j].URL
					item.Up = getStatus(item.DBName, item.SID)
					req.ReturnOK(w, r, item, 1)
					return
				}
			}
		}
		req.ReturnError(w, http.StatusNotFound, "not_found", "can't find url", err)
		return
	}

	// database mode
	err = ConfigDB.QueryRow(
		`SELECT url.url_id, url.db_id, db.name as db_name, url.sid, url.url 
		FROM url inner join db on url.db_id = db.db_id WHERE url_id = ?`,
		id).Scan(&item.ID, &item.DBId, &item.DBName, &item.SID, &item.URL)
	if errors.Is(err, sql.ErrNoRows) {
		req.ReturnError(w, http.StatusNotFound, "not_found", "can't find url", err)
		return
	}
	if err != nil {
		log.Error("Cannot read url", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read tbl", err)
		return
	}
	item.Up = getStatus(item.DBName, item.SID)
	req.ReturnOK(w, r, item, 1)
}

func urlGetManyHandler(w http.ResponseWriter, r *http.Request) {
	var urls []URL

	req := PrepareReq(w, r)
	m := ValuesToModifier(r.URL.Query(), URLColumns)

	// declarative mode
	if config.App.MapDatabase == "" {
		for i := range dbmap {
			for j := range dbmap[i].Urls {
				item := URL{
					ID:     dbmap[i].Urls[j].ID,
					DBId:   dbmap[i].ID,
					DBName: dbmap[i].Name,
					SID:    dbmap[i].Urls[j].SID,
					URL:    dbmap[i].Urls[j].URL,
					Up:     getStatus(dbmap[i].Name, dbmap[i].Urls[j].SID),
				}
				urls = append(urls, item)
			}
		}
		req.ReturnOK(w, r, urls, len(urls))
		return
	}

	// database mode
	query := BuildQuery(
		`SELECT url.url_id, url.db_id, db.name as db_name, url.sid, url.url 
		FROM url inner join db on url.db_id=db.db_id`,
		m)
	log.Debug("running query", "query", query, "modifier", m, "values", r.URL.Query())
	rows, err := ConfigDB.Query(query)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't read url list", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item URL
		err := rows.Scan(&item.ID, &item.DBId, &item.DBName, &item.SID, &item.URL)
		if err != nil {
			req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
			return
		}
		item.Up = getStatus(item.DBName, item.SID)
		item.Error = URLError[item.URL]
		urls = append(urls, item)
	}
	if err = rows.Err(); err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't scan item", err)
	}
	req.ReturnOK(w, r, urls, len(urls))
}

func urlPostOneHandler(w http.ResponseWriter, r *http.Request) {
	var item URL
	req := PrepareReq(w, r)

	// Fetch and decode body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("cannot read body: %v", err)
		req.ReturnError(w, http.StatusInternalServerError, "0000", "Cannot read request", err)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("could not decode url", "error", err)
		req.ReturnError(w, http.StatusBadRequest, "0003", "JSON parse error", err)
		return
	}
	if item.DBId == 0 || item.SID == "" || item.URL == "" {
		log.Error("missing parameters", "item", item)
		req.ReturnError(w, http.StatusBadRequest, "0003", "missing parameters", nil)
		return
	}
	// Add entry
	log.Debug("Creating db", "item", item)
	result, err := ConfigDB.Exec(
		`INSERT INTO url(db_id, sid, url) VALUES (?, ?, ?)`, item.DBId, item.SID, item.URL)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	item.ID, _ = result.LastInsertId()
	err = RefreshMappingTable()
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "0003", "cannot refresh mapping table", err)
		return
	}
	req.ReturnOK(w, r, item, 1)
}

func urlDeleteOneHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)

	id, err := ExtractID(r)
	if err != nil {
		log.Error("invalid id")
		req.ReturnError(w, http.StatusBadRequest, "invalid_id", "Invalid ID", err)
		return
	}

	result, err := ConfigDB.Exec(`DELETE FROM url WHERE url_id = ?`, id)
	if err != nil {
		log.Error("Cannot delete url", "id", id, "error", err)
		req.ReturnError(w, http.StatusInternalServerError, "SYSTEM", "can't delete url", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "url not found", nil)
		return
	}
	req.ReturnOK(w, r, nil, 0)
}

func urlPutOneHandler(w http.ResponseWriter, r *http.Request) {
	var item URL
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
	log.Debug("Updating url", "id", id, "item", item)
	// err = app.Validate.Struct(item)
	if item.DBId == 0 || item.SID == "" || item.URL == "" {
		log.Error("missing parameters", "item", item)
		req.ReturnError(w, http.StatusBadRequest, "0003", "missing parameters", nil)
		return
	}

	result, err := ConfigDB.Exec(
		`UPDATE url set sid=?, url=? where url_id=?`,
		item.SID, item.URL, id)
	if err != nil {
		req.ReturnError(w, http.StatusBadRequest, "0003", "Database error", err)
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		req.ReturnError(w, http.StatusNotFound, "NOT_FOUND", "url not found", nil)
		return
	}

	req.ReturnOK(w, r, item, 1)
}

func urlPostRestartAllHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	RootChannel <- "restart"

	// for i := range dbmap {
	// 	for j := range dbmap[i].Urls {
	// 		dbmap[i].Urls[j].commandChannel <- "restart"
	// 	}
	// }
	req.ReturnOK(w, r, nil, 0)
}
