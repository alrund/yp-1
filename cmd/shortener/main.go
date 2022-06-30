package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/handler"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.GetConfig()

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "Адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Базовый адрес результирующего сокращённого URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "Путь до файла с сокращёнными URL")
	flag.StringVar(&cfg.DatabaseDsn, "d", cfg.DatabaseDsn, "Строка с адресом подключения к БД")
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

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		handler.Ping(us, w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.Get(us, w, r)
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handler.GetUserURLs(us, w, r)
	}).Methods(http.MethodGet)

	r.Use(middleware.Compress)
	r.Use(middleware.Decompress)
	r.Use(middleware.Auth)

	log.Fatal(http.ListenAndServe(us.GetServerAddress(), r))
}
