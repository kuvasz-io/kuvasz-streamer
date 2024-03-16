package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	PackageName   string
	MaxGoroutines int
	log           *slog.Logger
}

var (
	app = App{
		MaxGoroutines: 100,
	}
)

func (app App) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	log := app.log
	log.Info("ReqNotFound",
		"method", r.Method,
		"uri", r.RequestURI,
		"remote_addr", r.RemoteAddr,
		"content_length", r.ContentLength,
		"headers", r.Header,
	)
	http.NotFound(w, r)
}

func StartAPI(log *slog.Logger) {
	// Set global logger
	app.log = log.With("module", "api")

	// Create HTTP Server
	router := mux.NewRouter().StrictSlash(true)

	// Add web admin and route by default
	router.Path("/").Handler(http.RedirectHandler("/admin/", http.StatusSeeOther))
	router.PathPrefix("/admin").Handler(http.FileServer(http.FS(webDist)))

	// Add utility handlers
	router.Path("/metrics").Handler(promhttp.Handler())
	router.NotFoundHandler = http.HandlerFunc(app.DefaultHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.DefaultHandler)

	// Start the engine
	log.Debug("Starting api server", "config", config.Server)
	srv := &http.Server{
		Handler:           router,
		Addr:              config.Server.Addr,
		ReadTimeout:       time.Duration(config.Server.ReadTimeout) * time.Second,
		ReadHeaderTimeout: time.Duration(config.Server.ReadHeaderTimeout) * time.Second,
		WriteTimeout:      time.Duration(config.Server.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(config.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes:    config.Server.MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
	log.Error("%v", srv.ListenAndServe())
	os.Exit(1)
}
