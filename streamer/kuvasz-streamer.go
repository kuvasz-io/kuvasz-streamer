package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // suppress linter error
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/time/rate"
)

const (
	// table types.
	TableTypeAppend  = "append"
	TableTypeHistory = "history"
	TableTypeClone   = "clone"
	StatusStarting   = "starting"
	StatusActive     = "active"
	StatusStopping   = "stopping"
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
	wg                 sync.WaitGroup
	Status             = StatusStarting
	lim                *rate.Limiter

	//go:embed migrations/*.sql
	embedMigrations embed.FS

	//go:embed admin
	webDist embed.FS
)

func SetStatus(s string) {
	log.Info("Setting status to", "status", s)
	Status = s
}

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

	// Start rate limiter
	lim = rate.NewLimiter(config.App.SyncRate, config.App.SyncBurst)
	_ = lim.Wait(context.Background()) // REMOVE ME
	// Start main loop
	RootChannel = make(chan string)
	for {
		SetStatus(StatusStarting)
		err := SetupDestination()
		if err != nil {
			log.Error("Error setting up destination", "err", err)
			os.Exit(1)
		}
		ReadMap()
		dbmap.CompileRegexes()
		// Create root context allowing cancellation of all goroutines
		rootContext, rootCancel := context.WithCancel(context.Background())

		// Loop through config and replicate databases
		log.Info("Start processing source databases")
		for _, database := range dbmap {
			for i, url := range database.Urls {
				log.Info("Starting replication thread", "db-sid", database.Name+"-"+url.SID, "url", url.URL)
				wg.Add(1)
				go DoReplicateDatabase(rootContext, database, &database.Urls[i])
			}
		}
		SetStatus(StatusActive)
		<-RootChannel
		rootCancel()
		SetStatus(StatusStopping)
		// wait until all workers exit
		log.Debug("Waiting for workers to exit")
		wg.Wait()
		CloseDestination()
		CloseConfigDB()
	}
}
