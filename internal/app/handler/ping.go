package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/alrund/yp-1/internal/app/config"
	_ "github.com/jackc/pgx/v4/stdlib" // pgx
)

type Configurator interface {
	GetConfig() *config.Config
}

func Ping(us Configurator, w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", us.GetConfig().DatabaseDsn)
	if err != nil {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
		return
	}
}
