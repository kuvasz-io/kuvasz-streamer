package main

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

func Migrate(embeddedMigrations embed.FS, directory string, db *sql.DB) {
	goose.SetBaseFS(embeddedMigrations)
	goose.SetLogger(GetLogger(log))

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, directory); err != nil {
		panic(err)
	}
}
