package main

import (
	"log"
	"net/http"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/handler"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/caarlos0/env/v6"
	"github.com/gorilla/mux"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080/"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// Инкремент 1
// Сервер должен быть доступен по адресу: http://localhost:8080.
// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201
// и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL
// и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
//
// Инкремент 2
// Покройте сервис юнит-тестами. Сконцентрируйтесь на покрытии тестами эндпоинтов,
// чтобы защитить API сервиса от случайных изменений.
//
// Инкремент 3
// Вы написали приложение с помощью стандартной библиотеки net/http.
// Используя любой пакет (роутер или фреймворк), совместимый с net/http, перепишите ваш код.
// Задача направлена на рефакторинг приложения с помощью готовой библиотеки.
// Обратите внимание, что необязательно запускать приложение вручную: тесты,
// которые вы написали до этого, помогут вам в рефакторинге.
//
// Инкремент 4
// Добавьте в сервер новый эндпоинт POST /api/shorten,
// принимающий в теле запроса JSON-объект {"url":"<some_url>"} и возвращающий в ответ объект {"result":"<shorten_url>"}.
//
// Инкремент 5
// Добавьте возможность конфигурировать сервис с помощью переменных окружения:
// - адрес запуска HTTP-сервера с помощью переменной SERVER_ADDRESS.
// - базовый адрес результирующего сокращённого URL с помощью переменной BASE_URL.
//
// Инкремент 6
// Сохраняйте все сокращённые URL на диск в виде файла. При перезапуске приложения все URL должны быть восстановлены.
// Путь до файла должен передаваться в переменной окружения FILE_STORAGE_PATH.
// При отсутствии переменной окружения или при её пустом значении вернитесь к хранению сокращённых URL в памяти.
func main() {
	var (
		err  error
		cfg  Config
		strg app.Storage = storage.NewMap()
	)

	if err = env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.FileStoragePath != "" {
		strg, err = storage.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	us := &app.URLShortener{
		ServerAddress:  cfg.ServerAddress,
		BaseURL:        cfg.BaseURL,
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

	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.Get(us, w, r)
	}).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(us.GetServerAddress(), r))
}
