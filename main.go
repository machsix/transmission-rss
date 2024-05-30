package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

var configPath string
var config atomic.Pointer[Config]
var runJob atomic.Bool

func main() {
	path := flag.String("path", "", "config dir path")
	rpc := flag.String("rpc", "http://127.0.0.1:9091/transmission/rpc", "transmission rpc url")
	lishost := flag.String("host", ":9093", "listen host")
	flag.Parse()

	updateConfig()

	cache, err := NewCacheByPath(filepath.Join(*path, "trss.db"))
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	tr, err := NewTransmission(*rpc)
	if err != nil {
		panic(err)
	}

	job := NewJob(tr, cache)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan func())

	go func() {
		if err := job.Start(ctx, ch, config.Load); err != nil {
			slog.Error("job start failed", "err", err)
		}
	}()

	go func() {
		if err := WatchConfig(ctx, filepath.Join(*path, "config.json"), func() {
			slog.Info("check config changed, reload config")
			updateConfig()
		}); err != nil {
			slog.Error("watch config failed", "err", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("GET /start_job", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if runJob.Load() {
			slog.Warn("job is already running")
			return
		}

		ch <- func() { runJob.Store(false) }
	}))

	if err := http.ListenAndServe(*lishost, mux); err != nil {
		panic(err)
	}
}

func updateConfig() {
	data, err := os.ReadFile(filepath.Join(configPath, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			cf := new(Config)
			config.Store(cf)

			data, err := json.Marshal(cf)
			if err != nil {
				slog.Error("marshal config failed", "err", err)
				return
			}

			err = os.MkdirAll(configPath, 0755)
			if err != nil {
				slog.Error("create config dir failed", "err", err)
				return
			}

			err = os.WriteFile(filepath.Join(configPath, "config.json"), data, 0644)
			if err != nil {
				slog.Error("write config failed", "err", err)
				return
			}

			return
		}
		slog.Error("read config failed", "err", err)
		return
	}

	cf := new(Config)
	err = json.Unmarshal(data, cf)
	if err != nil {
		slog.Error("unmarshal config failed", "err", err)
		return
	}

	config.Store(cf)
}
