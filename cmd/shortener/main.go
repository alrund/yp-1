package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/alrund/yp-1/internal/app/handler"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/gorilla/mux"
)

const defaultBuildValue string = "N/A"

var (
	buildVersion = defaultBuildValue
	buildDate    = defaultBuildValue
	buildCommit  = defaultBuildValue
)

func main() {
	printBuildInfo()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	cfg := config.GetConfig()

	us := &app.URLShortener{
		Config:         cfg,
		Storage:        getStorage(cfg),
		TokenGenerator: generator.NewSimple(),
	}

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           getRouter(us, cfg),
		ReadHeaderTimeout: 1 * time.Second,
	}

	if err := run(cfg, server, sigCh); err != nil {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func run(cfg *config.Config, server *http.Server, sigCh chan os.Signal) error {
	httpShutdownCh := make(chan struct{})

	go func() {
		<-sigCh

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(httpShutdownCh)
	}()

	log.Println("Starting HTTP server", cfg.ServerAddress)

	var err error
	if cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		err = server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		return err
	}

	<-httpShutdownCh
	fmt.Println("Server Shutdown gracefully")

	return nil
}

func getStorage(cfg *config.Config) app.Storage {
	var (
		err  error
		strg app.Storage = storage.NewMap()
	)

	if cfg.FileStoragePath != "" {
		strg, err = storage.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if cfg.DatabaseDsn != "" {
		strg, err = storage.NewDB(cfg.DatabaseDsn)
		if err != nil {
			log.Fatal(err)
		}
	}

	return strg
}

func getRouter(us *app.URLShortener, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.Add(us, w, r)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		handler.AddJSON(us, w, r)
	}).Methods(http.MethodPost)

	r.HandleFunc("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		handler.AddBatchJSON(us, w, r)
	}).Methods(http.MethodPost)

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		handler.Ping(us, w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.Get(us, w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handler.GetUserURLs(us, w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handler.DeleteURLs(us, w, r)
	}).Methods(http.MethodDelete)

	r.HandleFunc("/api/internal/stats", func(w http.ResponseWriter, r *http.Request) {
		handler.Stats(us, w, r)
	}).Methods(http.MethodGet)

	subRouter := r.PathPrefix("/debug/pprof").Subrouter()
	subRouter.HandleFunc("/", pprof.Index)
	subRouter.HandleFunc("/cmdline", pprof.Cmdline)
	subRouter.HandleFunc("/profile", pprof.Profile)
	subRouter.HandleFunc("/symbol", pprof.Symbol)
	subRouter.HandleFunc("/trace", pprof.Trace)
	subRouter.HandleFunc("/{name}", pprof.Index)

	r.Use(middleware.Compress)
	r.Use(middleware.Decompress)
	r.Use(middleware.Auth(encryption.NewEncryption(cfg.CipherPass)))

	return r
}
