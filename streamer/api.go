package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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

//nolint:wrapcheck // whole function is a wrapper
func (r *requestRecord) Write(p []byte) (int, error) {
	r.responseBytes += int64(len(p))
	return r.ResponseWriter.Write(p)
}

func (r *requestRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func AddCORSHeaders(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", config.Cors.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(config.Cors.AllowCredentials))
	w.Header().Set("Access-Control-Expose-Headers", config.Cors.AllowHeaders)
	w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.Cors.MaxAge))
}

func (app App) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	log := app.log
	log.Info("ReqNotFound",
		"method", r.Method,
		"uri", r.RequestURI,
		"remoteAddr", r.RemoteAddr,
		"contentLength", r.ContentLength,
		"headers", r.Header,
	)
	origin := r.Header.Get("Origin")
	AddCORSHeaders(w, origin)
	http.NotFound(w, r)
}

func ExtractID(r *http.Request) (int64, error) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		return 0, errors.New("missing id")
	}
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %w", err)
	}
	return i, nil
}

func PrepareReq(w http.ResponseWriter, r *http.Request) Request {
	// set default output type, can be overridden
	w.Header().Set("Content-Type", "application/json")

	return Request{
		Logger: log.With("url", r.RequestURI),
		User:   "admin",
	}
}

func (req Request) ReturnError(w http.ResponseWriter, status int, code string, message string, err error) {
	log := req.Logger

	w.WriteHeader(status)
	log.Error("request error", "status", status, "code", code, "message", message, "error", err)
	if err == nil {
		err = errors.New("no additional error information")
	}
	apiError := ErrorMessage{Code: code, Error: message, Info: err.Error()}
	jsonAPIError, err := json.Marshal(apiError)
	if err != nil {
		log.Error("can't marshal error message", "message", apiError, "error", err)
	}
	fmt.Fprint(w, string(jsonAPIError))
}

func (req Request) ReturnNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (req Request) ReturnSuccess(w http.ResponseWriter, _ *http.Request, code int, body any, count int) {
	log := req.Logger
	log.Debug("Handling success", "body", body)
	// No response body
	if body == nil {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(count))

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
	log := log.With("handler", "CORS")

	origin := r.Header.Get("Origin")
	log.Info("CORS Handler", "origin", origin)
	if !contains(origin, config.Cors.AllowedOrigins) && config.Cors.AllowedOrigins[0] != "*" {
		req := PrepareReq(w, r)
		log.Error("origin is not in the list of allowed origins", "origin", origin, "list", config.Cors.AllowedOrigins)
		req.ReturnError(w, 400, "origin_not_allowed", "", nil)
		return
	}

	AddCORSHeaders(w, origin)
	w.WriteHeader(http.StatusOK)
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
			AddCORSHeaders(w, origin)
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

func StatusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := log.With("handler", "Status")
		if Status != "active" && strings.HasPrefix(r.URL.Path, "/api") {
			req := PrepareReq(w, r)
			log.Error("Server is not ready", "status", Status)
			req.ReturnError(w, 400, "not_ready", "server not ready: "+Status, nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func DeclarativeModeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.App.MapDatabase == "" &&
			strings.HasPrefix(r.URL.Path, "/api") &&
			r.Method != http.MethodGet &&
			r.Method != http.MethodOptions {
			req := PrepareReq(w, r)
			req.ReturnError(w, 405, "not_allowed", "cannot modify configuration in declarative mode", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TokenValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip validation for non API requests
		if !strings.HasPrefix(r.URL.Path, "/api") && !strings.HasPrefix(r.URL.Path, "/refresh-token") {
			next.ServeHTTP(w, r)
			return
		}
		// skip validation for OPTIONS requests
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		// Validate token
		if _, ok := r.Header["Authorization"]; !ok {
			req := PrepareReq(w, r)
			req.ReturnError(w, http.StatusUnauthorized, "not_allowed", "no authorization header", nil)
			return
		}
		token := strings.TrimPrefix(r.Header["Authorization"][0], "Bearer ")
		role, err := validateToken(token)
		if err != nil {
			req := PrepareReq(w, r)
			req.ReturnError(w, http.StatusUnauthorized, "not_allowed", "invalid authorization token", nil)
			return
		}
		// Now check the allowed endpoints
		if role == "viewer" && r.Method != http.MethodGet {
			req := PrepareReq(w, r)
			req.ReturnError(w, http.StatusForbidden, "not_allowed", "viewer cannot modify configuration", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func APIServer(log *slog.Logger) {
	// Set global logger
	app.log = log.With("module", "api")

	// Create HTTP Server
	router := mux.NewRouter().StrictSlash(true)

	// Add middlewares
	router.Use(ObservabilityMiddleware)
	router.Use(StatusMiddleware)
	router.Use(DeclarativeModeMiddleware)
	router.Use(CORSMiddleware)

	// Add utility handlers
	router.Path("/metrics").Handler(promhttp.Handler())
	router.NotFoundHandler = http.HandlerFunc(app.DefaultHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.DefaultHandler)
	router.Methods("OPTIONS").HandlerFunc(CORSHandler)

	// Add app handlers
	router.HandleFunc("/api/map", mapGetManyHandler).Methods("GET")
	router.HandleFunc("/api/map/{id}", mapGetOneHandler).Methods("GET")
	router.HandleFunc("/api/map/{id}/create", mapCreateTableHandler).Methods("POST")
	router.HandleFunc("/api/map/{id}/clone", mapCloneTableHandler).Methods("POST")
	router.HandleFunc("/api/map/refresh", mapRefreshHandler).Methods("POST")

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
	router.HandleFunc("/api/url/restart", urlPostRestartAllHandler).Methods("POST")

	router.HandleFunc("/api/tbl/{id}", tblGetOneHandler).Methods("GET")
	router.HandleFunc("/api/tbl", tblGetManyHandler).Methods("GET")
	router.HandleFunc("/api/tbl", tblPostOneHandler).Methods("POST")
	router.HandleFunc("/api/tbl/{id}", tblDeleteOneHandler).Methods("DELETE")
	router.HandleFunc("/api/tbl/{id}", tblPutOneHandler).Methods("PUT")

	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/refresh-token", refreshTokenHandler).Methods("GET")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	// Add web admin and route by default
	router.Path("/").Handler(http.RedirectHandler("/admin/", http.StatusSeeOther))
	router.PathPrefix("/admin").Handler(http.FileServer(http.FS(webDist)))

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
