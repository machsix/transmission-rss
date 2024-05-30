package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"regexp"

	gtp "github.com/j-muller/go-torrent-parser"
	"github.com/mmcdole/gofeed"
)

/*
RSS feeds:
https://planet.openstreetmap.org/pbf/planet-pbf-rss.xml
https://planet.openstreetmap.org/pbf/full-history/history-pbf-rss.xml
https://planet.openstreetmap.org/planet/planet-bz2-rss.xml
https://planet.openstreetmap.org/planet/full-history/history-bz2-rss.xml
https://planet.openstreetmap.org/planet/changesets-bz2-rss.xml
https://planet.openstreetmap.org/planet/discussions-bz2-rss.xml
*/

func ParseString(data string) (*Channel, error) {
	ps := gofeed.NewParser()

	feed, err := ps.ParseString(data)
	if err != nil {
		return nil, err
	}

	return parse(feed), nil
}

func ParseUrl(url string) (*Channel, error) {
	ps := gofeed.NewParser()

	feed, err := ps.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return parse(feed), nil
}

func parse(feed *gofeed.Feed) *Channel {
	ch := &Channel{
		Title:    feed.Title,
		Url:      feed.Link,
		Describe: feed.Description,
	}

	for _, item := range feed.Items {
		if len(item.Enclosures) == 0 {
			continue
		}

		ch.Items = append(ch.Items, Item{
			Title:       item.Title,
			ContentType: item.Enclosures[0].Type,
			Url:         item.Enclosures[0].URL,
			Describe:    item.Description,
		})
	}

	return ch
}

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

	regexp        []*regexp.Regexp
	excludeRegexp []*regexp.Regexp
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
