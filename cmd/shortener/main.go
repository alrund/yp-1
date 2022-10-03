package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

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

	cfg := config.GetConfig()

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Базовый адрес результирующего сокращённого URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Путь до файла с сокращёнными URL")
	flag.StringVar(&cfg.DatabaseDsn, "d", cfg.DatabaseDsn, "Строка с адресом подключения к БД")
	flag.StringVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Использовать HTTPS")
	flag.StringVar(&cfg.CertFile, "crt", cfg.CertFile, "Файл с сертификатом")
	flag.StringVar(&cfg.KeyFile, "key", cfg.KeyFile, "Файл с приватным ключом")
	flag.Parse()

	var err error
	var strg app.Storage = storage.NewMap()
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

	us := &app.URLShortener{
		Config:         cfg,
		Storage:        strg,
		TokenGenerator: generator.NewSimple(),
	}

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

	//nolint
	if cfg.EnableHTTPS != "" && cfg.CertFile != "" && cfg.KeyFile != "" {
		log.Fatal(http.ListenAndServeTLS(us.GetServerAddress(), cfg.CertFile, cfg.KeyFile, r))
	} else {
		log.Fatal(http.ListenAndServe(us.GetServerAddress(), r))
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
