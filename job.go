package main

import (
	"context"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
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
			runJob.Store(true)
			j.Do(getConfig)
			runJob.Store(false)
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

	m := splitConfigByHostname(config)

	wg := &sync.WaitGroup{}

	for _, v := range m {
		wg.Add(1)

		go func() {
			defer wg.Done()
			j.DoOne(v)
		}()
	}

	wg.Wait()

	slog.Info("job done")
}

func (j *Job) DoOne(config *Config) {
	type Result struct {
		channels []Channel
		rss      *RSS
	}
	ch := make(chan Result, 10)

	go func() {
		wg := &sync.WaitGroup{}
		semaphore := semaphore.NewWeighted(15)
		defer close(ch)
		for _, v := range config.Rss {
			if v.ExpiredOrDisabled() {
				continue
			}

			time.Sleep(time.Millisecond * 100)

			wg.Add(1)

			_ = semaphore.Acquire(context.Background(), 1)
			go func() {
				defer wg.Done()
				defer semaphore.Release(1)
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				chs, err := ParseUrl(ctx, v.Url)
				cancel()
				if err != nil {
					slog.Error("parse rss failed", "err", err, "url", v.Url, "name", v.Name)
					return
				}

				ch <- Result{
					channels: chs,
					rss:      v,
				}
			}()
		}
		wg.Wait()
	}()

	for r := range ch {
		chs := r.channels
		v := r.rss

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

				if v.FetchInterval > 0 {
					time.Sleep(time.Duration(v.FetchInterval) * time.Millisecond)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				tr, err := item.Get(ctx)
				cancel()
				if err != nil {
					slog.Error("get torrent failed", "err", err, "url", item.Url, "name", v.Name)
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

func splitConfigByHostname(config *Config) map[string]*Config {
	m := make(map[string]*Config)
	for _, v := range config.Rss {
		uri, err := url.Parse(v.Url)
		if err != nil {
			m["default"] = config
			continue
		}

		x := m[uri.Host]
		if x == nil {
			x = new(Config)
			m[uri.Host] = x
		}

		x.Rss = append(x.Rss, v)
	}
	return m
}
