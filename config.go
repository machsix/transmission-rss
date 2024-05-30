package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	gtp "github.com/j-muller/go-torrent-parser"
)

type Channel struct {
	Title    string
	Url      string
	Describe string

	Items []Item
}

type Item struct {
	Title       string
	ContentType string
	Url         string
	Describe    string
	PubDate     time.Time
}

func (i *Item) Get(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", i.Url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", i.ContentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

type RSS struct {
	Name          string   `json:"name,omitempty"`
	Url           string   `json:"url,omitempty"`
	DownloadDir   string   `json:"download_dir,omitempty"`
	Internal      int      `json:"internal,omitempty"`
	Regexp        []string `json:"regexp,omitempty"`
	ExcludeRegexp []string `json:"exclude_regexp,omitempty"`
	DownloadAfter int64    `json:"download_after,omitempty"`

	regexp        []*regexp.Regexp
	excludeRegexp []*regexp.Regexp
	downloadAfter time.Time
}

func (r *RSS) MatchDate(pubDate time.Time) bool {
	if r.DownloadAfter == 0 {
		return true
	}

	if r.downloadAfter.IsZero() {
		r.downloadAfter = time.Unix(r.DownloadAfter, 0)
	}

	return pubDate.After(r.downloadAfter)
}

func (r *RSS) Match(title string) bool {
	if r.regexp == nil && len(r.Regexp) > 0 {
		regexps := make([]*regexp.Regexp, 0, len(r.Regexp))

		for _, v := range r.Regexp {
			rr, err := regexp.Compile(v)
			if err != nil {
				slog.Error("compile regexp failed", "err", err, "regexp", v)
				continue
			}

			regexps = append(regexps, rr)

		}

		r.regexp = regexps
	}

	if r.excludeRegexp == nil && len(r.ExcludeRegexp) > 0 {
		regexps := make([]*regexp.Regexp, 0, len(r.ExcludeRegexp))

		for _, v := range r.ExcludeRegexp {
			rr, err := regexp.Compile(v)
			if err != nil {
				slog.Error("compile regexp failed", "err", err, "regexp", v)
				continue
			}

			regexps = append(regexps, rr)
		}

		r.excludeRegexp = regexps
	}

	for _, v := range r.excludeRegexp {
		if v.MatchString(title) {
			return false
		}
	}

	if len(r.Regexp) == 0 {
		return true
	}

	for _, v := range r.regexp {
		if v.MatchString(title) {
			return true
		}
	}

	return false
}

type Config struct {
	Rss []*RSS `json:"rss,omitempty"`
}

type Torrent struct {
	Torrent *gtp.Torrent
	Bytes   []byte
}

func ParseTorrent(data []byte) (*Torrent, error) {
	gt, err := gtp.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &Torrent{
		Torrent: gt,
		Bytes:   data,
	}, nil
}
