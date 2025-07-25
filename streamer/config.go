package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
	"golang.org/x/time/rate"
)

type (
	ServerConfig struct {
		Name              string `koanf:"name"`
		Addr              string `koanf:"address"`
		MaxGoroutines     int    `koanf:"max_goroutines"`
		ReadTimeout       int    `koanf:"read_timeout"`
		ReadHeaderTimeout int    `koanf:"read_header_timeout"`
		WriteTimeout      int    `koanf:"write_timeout"`
		IdleTimeout       int    `koanf:"idle_timeout"`
		MaxHeaderBytes    int    `koanf:"max_header_bytes"`
	}

	MaintenanceConfig struct {
		Pprof      string `koanf:"pprof"`
		StartDelay int    `koanf:"start_delay"`
	}
	DatabaseConfig struct {
		URL    string `koanf:"url"`
		Origin string `koanf:"origin"`
		Schema string `koanf:"schema"`
	}

	AuthConfig struct {
		AdminPassword string `koanf:"admin_password"`
		JWTKey        string `koanf:"jwt_key"`
		TTL           int    `koanf:"ttl"`
	}
	AppConfig struct {
		MapFile       string     `koanf:"map_file"`
		MapDatabase   string     `koanf:"map_database"`
		NumWorkers    int        `koanf:"num_workers"`
		CommitDelay   float64    `koanf:"commit_delay"`
		DefaultSchema string     `koanf:"default_schema"`
		SyncRate      rate.Limit `koanf:"sync_rate"`
		SyncBurst     int        `koanf:"sync_burst"`
	}

	CORSConfig struct {
		AllowedOrigins   []string `koanf:"allowed_origins"`
		AllowMethods     string   `koanf:"allow_methods"`
		AllowHeaders     string   `koanf:"allow_headers"`
		AllowCredentials bool     `koanf:"allow_credentials"`
		MaxAge           int      `koanf:"max_age"`
	}
	Config struct {
		Server      ServerConfig      `koanf:"server"`
		Maintenance MaintenanceConfig `koanf:"maintenance"`
		Logs        LogsConfig        `koanf:"logs"`
		Database    DatabaseConfig    `koanf:"database"`
		App         AppConfig         `koanf:"app"`
		Auth        AuthConfig        `koanf:"auth"`
		Cors        CORSConfig        `koanf:"cors"`
	}
)

var config = Config{
	Server: ServerConfig{
		Name:              "kuvasz-streamer",
		Addr:              ":8000",
		MaxGoroutines:     100,
		ReadTimeout:       30,
		ReadHeaderTimeout: 30,
		WriteTimeout:      30,
		IdleTimeout:       30,
		MaxHeaderBytes:    1000,
	},
	Maintenance: MaintenanceConfig{
		Pprof:      "",
		StartDelay: 0,
	},
	Logs: defaultLogsConfig,
	Database: DatabaseConfig{
		URL:    "",
		Origin: "",
		Schema: "public",
	},
	App: AppConfig{
		MapFile:       "/etc/kuvasz/map.yaml",
		MapDatabase:   "",
		NumWorkers:    2,
		CommitDelay:   1.0,
		DefaultSchema: "public",
		SyncRate:      1_000_000_000,
		SyncBurst:     1_000,
	},
	Auth: AuthConfig{
		AdminPassword: "$2b$05$KlJx0xWATjLt84bXrg6uZe/zU4TH3TvbPDLf6tOrzMUPEyN7AoEie",
		JWTKey:        "Y3OYHx7Y1KsRJPzJKqHGWfEaHsPbmwwSpPrXcND95Pw=",
		TTL:           300,
	},
	Cors: CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowMethods:     "GET,POST,PATCH,PUT,DELETE",
		AllowHeaders:     "Authorization,User-Agent,If-Modified-Since,Cache-Control,Content-Type,X-Total-Count",
		AllowCredentials: true,
		MaxAge:           86400,
	},
}

var k = koanf.New(".")

func reloadConfig(configFiles []string, f *flag.FlagSet, envPrefix string) {
	log.Warn("Loading config")
	k = koanf.New(".")
	for _, c := range configFiles {
		log.Warn("Parsing config file", "name", c)
		f := file.Provider(c)
		if err := k.Load(f, toml.Parser()); err != nil {
			log.Warn("skipping file", "name", c, "error", err)
		}
	}
	log.Warn("config file", "map", k.Sprint())
	// Step 2 - Override with environment variables
	log.Debug("Merging environment variables", "prefix", envPrefix)
	err := k.Load(env.Provider(envPrefix+"_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, envPrefix+"_")), "_", ".")
	}), nil)
	if err != nil {
		log.Error("can't load environment", "error", err)
	}
	log.Warn("after env variables", "map", k.Sprint())

	// Step 3 - Override with command line
	log.Warn("Merging command line flags")
	if err = k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		log.Error("error loading config", "error", err)
		os.Exit(1)
	}
	log.Warn("after command line", "map", k.Sprint())

	// Step 4 - Unmarshal into struct
	err = k.Unmarshal("", &config)
	if err != nil {
		log.Error("can't unmarshal config", "error", err)
	}
}

func Configure(configFiles []string, envPrefix string) {
	SetupLogs(defaultLogsConfig)

	flags := flag.NewFlagSet("config", flag.ContinueOnError)
	flags.Usage = func() {
		//nolint:forbidigo // Allow printing usage
		fmt.Printf("%s version %s built %s\n", Package, Version, Build)
		//nolint:forbidigo // Allow printing usage
		fmt.Println(flags.FlagUsages())
		os.Exit(1)
	}
	// Path to one or more config files to load into koanf along with some config params.
	flags.StringSlice("conf", configFiles, "path to one or more .toml config files")
	flags.String("server.name", "kuvasz-streamer", "service name")
	flags.String("server.address", ":8000", "service listen address")
	flags.Bool("logs.source", false, "print source line")
	flags.String("logs.level", "debug", "logging level")
	flags.String("database.url", "", "database connection string")
	flags.String("database.origin", "", "replication origin")
	flags.String("app.map", "map.yaml", "mapping file")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		//nolint:forbidigo // Allow printing usage
		fmt.Printf("Can't parse flags: %v\n", err)
		os.Exit(1)
	}

	// Load the config files provided in the command line.
	configFileNames, _ := flags.GetStringSlice("conf")
	for _, c := range configFileNames {
		log.Debug("Watching file", "name", c)
		f := file.Provider(c)
		err = f.Watch(func(_ interface{}, _ error) {
			reloadConfig(configFileNames, flags, envPrefix)
		})
		if err != nil {
			log.Error("can't watch file", "name", c, "error", err)
		}
	}
	reloadConfig(configFileNames, flags, envPrefix)
}
