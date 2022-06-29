package handler

import (
	"context"
	"database/sql"
	"github.com/alrund/yp-1/internal/app/config"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Configurator interface {
	GetConfig() *config.Config
}

func Ping(us Configurator, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

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
