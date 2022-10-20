package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
)

type Counter interface {
	GetConfig() *config.Config
	GetStats() (*app.Stat, error)
}

// Stats get statistic information.
func Stats(us Counter, w http.ResponseWriter, r *http.Request) {
	cfg := us.GetConfig()
	if cfg.TrustedSubnet == "" {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	realIPHeader := r.Header.Get("X-Real-IP")
	if realIPHeader == "" {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	realIP := net.ParseIP(realIPHeader)

	if !ipnet.Contains(realIP) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	stat, err := us.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(stat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
