package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	App struct {
		PackageName   string
		MaxGoroutines int
		log           *slog.Logger
	}
	ErrorMessage struct {
		Code  string `json:"code"`            // error_code
		Error string `json:"error,omitempty"` // error message
		Info  any    `json:"info,omitempty"`  // Additional information - error dependent
	}
	Request struct {
		Logger      *slog.Logger
		User        string
		Permissions any
	}

	RequestKey struct{}

	requestRecord struct {
		http.ResponseWriter
		status        int
		responseBytes int64
	}
)

var (
	app = App{
		MaxGoroutines: 100,
	}
)

func (r *requestRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *requestRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (app App) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	log := app.log
	log.Info("ReqNotFound",
		"method", r.Method,
		"uri", r.RequestURI,
		"remote_addr", r.RemoteAddr,
		"content_length", r.ContentLength,
		"headers", r.Header,
	)
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", config.Cors.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Expose-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.Cors.MaxAge))
	http.NotFound(w, r)
}

func ExtractID(r *http.Request) (int64, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return 0, fmt.Errorf("missing id")
	}
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func PrepareReq(w http.ResponseWriter, r *http.Request) (req Request) {
	// set default output type, can be overridden
	w.Header().Set("Content-Type", "application/json")

	log.Debug("Handling request", "req", req)
	return req
}

func (req Request) ReturnError(w http.ResponseWriter, status int, code string, message string, err error) {
	w.WriteHeader(status)
	apiError := ErrorMessage{Code: code, Error: message, Info: err.Error()}
	jsonAPIError, err := json.Marshal(apiError)
	if err != nil {
		req.Logger.Error("can't marshal error message", "message", apiError, "error", err)
	}
	fmt.Fprint(w, string(jsonAPIError))
}

func (req Request) ReturnNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (req Request) ReturnSuccess(w http.ResponseWriter, _ *http.Request, code int, body any, count int) {
	log.Debug("Handling success", "body", body)
	// No response body
	if body == nil {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("X-Total-Count", fmt.Sprint(count))

	// Convert response to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		req.ReturnError(w, http.StatusInternalServerError, "json_error", "Can't marshal response body", err)
		return
	}
	w.WriteHeader(code)
	fmt.Fprint(w, string(jsonBody))
}

func (req Request) ReturnOK(w http.ResponseWriter, r *http.Request, body interface{}, count int) {
	req.ReturnSuccess(w, r, http.StatusOK, body, count)
}

func (req Request) ReturnCreated(w http.ResponseWriter, r *http.Request, body interface{}, count int) {
	req.ReturnSuccess(w, r, http.StatusCreated, body, count)
}

func (req Request) ReturnAccepted(w http.ResponseWriter, r *http.Request, body interface{}, count int) {
	req.ReturnSuccess(w, r, http.StatusAccepted, body, count)
}

func contains(item string, list []string) bool {
	for _, listItem := range list {
		if item == listItem {
			return true
		}
	}
	return false
}

func CORSHandler(w http.ResponseWriter, r *http.Request) {
	req := PrepareReq(w, r)
	log := log.With("handler", "CORS")

	origin := r.Header.Get("Origin")
	log.Info("CORS Handler", "Origin", origin)
	if !contains(origin, config.Cors.AllowedOrigins) && config.Cors.AllowedOrigins[0] != "*" {
		log.Error("origin is not in the list of allowed origins", "origin", origin, "list", config.Cors.AllowedOrigins)
		req.ReturnError(w, 400, "origin_not_allowed", "", nil)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", config.Cors.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Expose-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.Cors.MaxAge))
	w.WriteHeader(200)
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		log := log.With("Middleware", "CORS")
		if origin != "" {
			log.Debug("Checking Origin", "origin", origin)
			if !contains(origin, config.Cors.AllowedOrigins) && config.Cors.AllowedOrigins[0] != "*" {
				log.Error("origin is not in the list of allowed origins", "origin", origin, "allowed origins", config.Cors.AllowedOrigins)
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", config.Cors.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", config.Cors.AllowHeaders)
			w.Header().Set("Access-Control-Expose-Headers", config.Cors.AllowHeaders)
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.Cors.MaxAge))
		}
		next.ServeHTTP(w, r)
	})
}

func ObservabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := &requestRecord{
			ResponseWriter: w,
			status:         http.StatusOK,
			responseBytes:  0,
		}

		// Log request
		log.Info("ReqStart",
			"method", r.Method,
			"uri", r.RequestURI,
			"remote_addr", r.RemoteAddr,
			"content_length", r.ContentLength,
			"headers", r.Header,
		)

		// Pass downstream
		start := time.Now()
		next.ServeHTTP(record, r)
		duration := time.Since(start).Milliseconds()

		// Log response
		log.Info("ReqStop",
			"method", r.Method,
			"uri", r.RequestURI,
			"remote_addr", r.RemoteAddr,
			"headers", record.ResponseWriter.Header(),
			"status", record.status,
			"request_size", r.ContentLength,
			"response_size", record.responseBytes,
			"time", duration,
		)
	})
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
	router.Methods("OPTIONS").HandlerFunc(CORSHandler)

	// Add app handlers
	router.HandleFunc("/api/db/{id}", dbGetOneHandler).Methods("GET")
	router.HandleFunc("/api/db", dbGetManyHandler).Methods("GET")
	router.HandleFunc("/api/db", dbPostOneHandler).Methods("POST")
	router.HandleFunc("/api/db/{id}", dbDeleteOneHandler).Methods("DELETE")
	router.HandleFunc("/api/db/{id}", dbPutOneHandler).Methods("PUT")

	router.HandleFunc("/api/url/{id}", urlGetOneHandler).Methods("GET")
	router.HandleFunc("/api/url", urlGetManyHandler).Methods("GET")
	router.HandleFunc("/api/url", urlPostOneHandler).Methods("POST")
	router.HandleFunc("/api/url/{id}", urlDeleteOneHandler).Methods("DELETE")
	router.HandleFunc("/api/url/{id}", urlPutOneHandler).Methods("PUT")

	router.HandleFunc("/api/tbl/{id}", tblGetOneHandler).Methods("GET")
	router.HandleFunc("/api/tbl", tblGetManyHandler).Methods("GET")
	router.HandleFunc("/api/tbl", tblPostOneHandler).Methods("POST")
	router.HandleFunc("/api/tbl/{id}", tblDeleteOneHandler).Methods("DELETE")
	router.HandleFunc("/api/tbl/{id}", tblPutOneHandler).Methods("PUT")

	// Add middlewares
	router.Use(ObservabilityMiddleware)
	router.Use(CORSMiddleware)

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
