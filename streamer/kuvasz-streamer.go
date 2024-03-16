package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // suppress linter error
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// table types.
	TableTypeAppend  = "append"
	TableTypeHistory = "history"
	TableTypeClone   = "clone"
)

var (
	Package            string
	Version            string
	Build              string
	DestConnectionPool *pgxpool.Pool
	err                error
	dbmap              DBMap
	destTables         PGTables
	ConfigDB           *sql.DB

	//go:embed migrations/*.sql
	embedMigrations embed.FS

	//go:embed admin
	webDist embed.FS
)

func main() {
	Configure(
		[]string{"./conf/kuvasz-streamer.toml", "/etc/kuvasz/kuvasz-streamer.toml"},
		"KUVASZ",
	)
	SetupLogs(config.Logs)
	log.Debug("Starting...")

	// Start pprof if configured
	if config.Maintenance.Pprof != "" {
		go func() {
			//nolint:forbidigo,gosec // pprof requires this
			fmt.Println(http.ListenAndServe(config.Maintenance.Pprof, nil))
		}()
	}

	// Connect to target database
	DestConnectionPool, err = pgxpool.New(context.Background(), config.Database.URL)
	if err != nil {
		log.Error("Can't connect to target database", "url", config.Database.URL, "error", err)
		os.Exit(1)
	}
	log.Info("Connected to target database", "url", config.Database.URL)

	// Get destination metadata
	destTables, err = GetTables(log, DestConnectionPool, "public")
	if err != nil {
		log.Error("Can't get destination table metadata", "error", err)
		os.Exit(1)
	}

	if config.App.MapDatabase != "" {
		ConfigDB, err = sql.Open("sqlite3", config.App.MapDatabase)
		if err != nil {
			log.Error("Can't open map database", "database", config.App.MapFile, "error", err)
			os.Exit(1)
		}
		Migrate(embedMigrations, "migrations", ConfigDB)
		ReadMapDatabase(ConfigDB)
	} else {
		ReadMapFile(config.App.MapFile)
	}
	CompileRegexes()

	// Start destination processing worker routines
	StartWorkers(config.App.NumWorkers)

	// Loop through config and replicate databases
	log.Info("Start processing source databases")
	for _, database := range dbmap {
		for _, url := range database.Urls {
			log.Info("Starting replication thread", "db", database.Name, "url", url.URL, "id", url.SID)
			go DoReplicateDatabase(database, url)
		}
	}
	// Start API Server
	StartAPI(log)
}