package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // suppress linter error
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v2"
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
)

func main() {
	Configure(
		[]string{"kuvasz-streamer.toml", "./conf/kuvasz-streamer.toml", "/etc/kuvasz-streamer.toml"},
		"KUVASZ",
	)
	SetupLogs(config.Logs)
	log.Debug("Starting...")

	// Start pprof if configured
	if config.Server.Pprof != "" {
		go func() {
			//nolint:forbidigo,gosec // pprof requires this
			fmt.Println(http.ListenAndServe(config.Server.Pprof, nil))
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

	// Read map
	log.Info("Reading map file", "filename", config.App.MapFile)
	var data []byte
	data, err = os.ReadFile(config.App.MapFile)
	if err != nil {
		log.Error("Can't read map file", "filename", config.App.MapFile, "error", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(data, &dbmap)
	if err != nil {
		log.Error("Can't unmarshal map file", "filename", config.App.MapFile, "error", err)
	}
	log.Info(fmt.Sprintf("Map config: %v", dbmap))

	// Loop throught config and replicate databases
	log.Info("Start processing source databases")
	for _, database := range dbmap {
		for _, url := range database.Urls {
			log.Info("Starting replication thread", "db", database.Name, "url", url.URL, "id", url.SID)
			tables := database.Tables
			go DoReplicateDatabase(url.URL, url.SID, tables)
		}
	}
	// block forever
	select {}
}
