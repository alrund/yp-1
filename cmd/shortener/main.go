package main

import (
	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/handler"
	stg "github.com/alrund/yp-1/internal/app/storage"
	gen "github.com/alrund/yp-1/internal/app/token/generator"
	"log"
	"net/http"
)

// Инкремент 1
// Сервер должен быть доступен по адресу: http://localhost:8080.
// Сервер должен предоставлять два эндпоинта: POST / и GET /{id}.
// Эндпоинт POST / принимает в теле запроса строку URL для сокращения и возвращает ответ с кодом 201 и сокращённым URL в виде текстовой строки в теле.
// Эндпоинт GET /{id} принимает в качестве URL-параметра идентификатор сокращённого URL и возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
// Нужно учесть некорректные запросы и возвращать для них ответ с кодом 400.
//
// Инкремент 2
//Покройте сервис юнит-тестами. Сконцентрируйтесь на покрытии тестами эндпоинтов, чтобы защитить API сервиса от случайных изменений.
func main() {
	us := &app.URLShortener{
		Schema:         "http",
		Host:           "localhost:8080",
		Storage:        stg.NewMapStorage(),
		TokenGenerator: gen.NewSimpleGenerator(),
	}
	err := http.ListenAndServe(us.GetServerHost(), http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handler.AddHandler(us, w, r)
			case http.MethodGet:
				handler.GetHandler(us, w, r)
			default:
				http.Error(w, "Only GET & POST requests are allowed!", http.StatusMethodNotAllowed)
			}
		},
	))
	log.Fatal(err)
}
