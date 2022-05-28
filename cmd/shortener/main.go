package main

import (
	"log"
	"net/http"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/handler"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token/generator"
	"github.com/gorilla/mux"
)

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
func main() {
	us := &app.URLShortener{
		Schema:         "http",
		Host:           "localhost:8080",
		Storage:        storage.NewMap(),
		TokenGenerator: generator.NewSimple(),
	}

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.Add(us, w, r)
	}).Methods(http.MethodPost)
	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		handler.Get(us, w, r)
	}).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(us.GetServerHost(), r))
}
