package main

import (
	"log/slog"
	"net/http"
)

func route(mux *http.ServeMux) {
	mux.Handle("PATCH /job/disabled", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		configMu.Lock()
		defer configMu.Unlock()

		cf := config.Load()

		for _, v := range cf.Rss {
			if v.Name == name {
				v.Disabled = !v.Disabled
			}
		}

		if err := saveConfig(cf); err != nil {
			slog.Error("save config failed", "err", err)
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(200)
	}))
}
