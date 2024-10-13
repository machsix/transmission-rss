package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hekmon/transmissionrpc/v3"
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

func (i *Item) Get(ctx context.Context) (Torrent, error) {
	if strings.HasPrefix(i.Url, "magnet:?xt=") {
		return TorrentHash(i.Url), nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", i.Url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", i.ContentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	tr, err := ParseTorrent(data)
	if err != nil {
		return nil, fmt.Errorf("parse torrent failed: %w, data: %s", err, data)
	}

	return tr, nil
}

type RSS struct {
	Disabled      bool     `json:"disabled,omitempty" toml:"disabled"`
	Name          string   `json:"name,omitempty" toml:"name"`
	Url           string   `json:"url,omitempty" toml:"url"`
	DownloadDir   string   `json:"download_dir,omitempty" toml:"download_dir"`
	Internal      int      `json:"internal,omitempty" toml:"internal"`
	Regexp        []string `json:"regexp,omitempty" toml:"regexp"`
	ExcludeRegexp []string `json:"exclude_regexp,omitempty" toml:"exclude_regexp"`
	DownloadAfter int64    `json:"download_after,omitempty" toml:"download_after"`
	ExpireTime    int64    `json:"expire_time,omitempty" toml:"expire_time"`
	FetchInterval int64    `json:"fetch_interval,omitempty" toml:"fetch_interval"`

	regexp        []*regexp.Regexp
	excludeRegexp []*regexp.Regexp
	downloadAfter time.Time
	expireTime    time.Time
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
func (r *RSS) ExpiredOrDisabled() bool {
	if r.Disabled {
		return true
	}

	if r.ExpireTime == 0 {
		return false
	}

	if r.ExpireTime != 0 && r.expireTime.IsZero() {
		r.expireTime = time.Unix(r.ExpireTime, 0)
	}

	return r.expireTime.Before(time.Now())
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
	Rss []*RSS `json:"rss,omitempty" toml:"rss"`
}

type Torrent interface {
	AddPayload(downloadDir string) transmissionrpc.TorrentAddPayload
}

type TorrentHash string

func (th TorrentHash) AddPayload(downloadDir string) transmissionrpc.TorrentAddPayload {
	return transmissionrpc.TorrentAddPayload{
		DownloadDir: &downloadDir,
		Filename:    (*string)(&th),
	}
}

type TorrentFile struct {
	Torrent *gtp.Torrent
	Bytes   []byte
}

func (tr *TorrentFile) AddPayload(downloadDir string) transmissionrpc.TorrentAddPayload {
	str := base64.StdEncoding.EncodeToString(tr.Bytes)
	return transmissionrpc.TorrentAddPayload{
		DownloadDir: &downloadDir,
		MetaInfo:    &str,
	}
}

func ParseTorrent(data []byte) (*TorrentFile, error) {
	gt, err := gtp.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &TorrentFile{
		Torrent: gt,
		Bytes:   data,
	}, nil
}
