package main

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	tr    *Transmission
	cache Cache
}

func NewJob(tr *Transmission, cache Cache) *Job {
	return &Job{
		tr:    tr,
		cache: cache,
	}
}

func (j *Job) Start(ctx context.Context, notify chan func(), getConfig func() *Config) error {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			j.Do(getConfig)
		case f := <-notify:
			j.Do(getConfig)
			ticker.Reset(time.Hour)
			if f != nil {
				f()
			}
		}
	}
}

func (j *Job) Do(getConfig func() *Config) {
	config := getConfig()

	for _, v := range config.Rss {
		chs, err := ParseUrl(context.TODO(), v.Url)
		if err != nil {
			slog.Error("parse rss failed", "err", err, "url", v.Url, "name", v.Name)
			continue
		}

		for _, ch := range chs {
			for _, item := range ch.Items {
				if !v.MatchDate(item.PubDate) {
					continue
				}

				if !v.Match(item.Title) {
					continue
				}

				_, ok := j.cache.Load(v.Url, item.Url)
				if ok {
					continue
				}

				data, err := item.Get(context.Background())
				if err != nil {
					slog.Error("get torrent failed", "err", err, "url", item.Url, "name", v.Name)
					continue
				}

				tr, err := ParseTorrent(data)
				if err != nil {
					slog.Error("parse torrent failed", "err", err, "url", item.Url, "name", v.Name)
					continue
				}

				err = j.tr.Add(context.TODO(), tr, v.DownloadDir)
				if err != nil {
					slog.Error("add torrent failed", "err", err, "url", item.Url, "name", v.Name)
					continue
				}

				slog.Info("add torrent", "url", item.Url, "name", item.Title)

				err = j.cache.Store(v.Url, item.Url, tr)
				if err != nil {
					slog.Error("store torrent failed", "err", err, "url", item.Url, "name", v.Name)
					continue
				}
			}
		}
	}
}
