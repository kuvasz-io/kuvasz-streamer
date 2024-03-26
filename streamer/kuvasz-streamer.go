package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // suppress linter error
	"time"

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
	ConfigDB           *sql.DB
	DestConnectionPool *pgxpool.Pool
	DestTables         PGTables
	dbmap              DBMap
	URLError           = make(map[string]string)
	RootChannel        chan string

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

	// Start destination processing worker routines
	StartWorkers(config.App.NumWorkers)

	// Start API Server
	go APIServer(log)

	// Start main loop
	RootChannel = make(chan string)
	for {
		SetupDestination()
		ReadMap()
		CompileRegexes()
		// Create root context allowing cancellation of all goroutines
		rootContext, rootCancel := context.WithCancel(context.Background())

		// Loop through config and replicate databases
		log.Info("Start processing source databases")
		for _, database := range dbmap {
			for i, url := range database.Urls {
				log.Info("Starting replication thread", "db", database.Name, "url", url.URL, "sid", url.SID)
				go DoReplicateDatabase(rootContext, database, &database.Urls[i])
			}
		}
		<-RootChannel
		rootCancel()
		// wait until all workers exit
		time.Sleep(1 * time.Second)
		CloseDestination()
		CloseConfigDB()
	}
}
