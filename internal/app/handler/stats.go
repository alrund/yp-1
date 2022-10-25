package handler

import (
	"encoding/json"
	"net"
	"net/http"
)

// Stats get statistic information.
func (hc *Collection) Stats() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cfg := hc.us.GetConfig()
		if cfg.TrustedSubnet == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if hc.us.TrustedSubnet == nil {
			_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			hc.us.TrustedSubnet = ipnet
		}

		realIPHeader := r.Header.Get("X-Real-IP")
		if realIPHeader == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		realIP := net.ParseIP(realIPHeader)

		if !hc.us.TrustedSubnet.Contains(realIP) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		stat, err := hc.us.GetStats()
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
	return fn
}
