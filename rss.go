package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
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

func ParseString(data string) ([]Channel, error) {
	var rf RSSFeed
	err := xml.Unmarshal([]byte(data), &rf)
	if err != nil {
		return nil, err
	}

	var chs []Channel

	for _, channel := range rf.Channel {
		ch := Channel{
			Title:    channel.Title,
			Url:      channel.Link,
			Describe: channel.Description,
		}

		for _, item := range channel.Items {
			if len(item.Enclosure) == 0 {
				continue
			}

			it := Item{
				Title:       item.Title,
				ContentType: item.Enclosure[0].Type,
				Url:         item.Enclosure[0].URL,
				Describe:    item.Description,
			}

			if item.PubDate != "" {
				it.PubDate = ParseTime(item.PubDate)
			} else if item.Torrent.PubDate != "" {
				it.PubDate = ParseTime(item.Torrent.PubDate)
			}

			ch.Items = append(ch.Items, it)
		}

		chs = append(chs, ch)
	}

	return chs, nil
}

func ParseUrl(ctx context.Context, url string) ([]Channel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// if f.AuthConfig != nil && f.AuthConfig.Username != "" && f.AuthConfig.Password != "" {
	// 	req.SetBasicAuth(f.AuthConfig.Username, f.AuthConfig.Password)
	// }

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseString(string(data))
}

type RSSFeed struct {
	Version string           `xml:"version,attr"`
	Channel []RssFeedChannel `xml:"channel"`
}

type RssFeedChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Copyright   string    `xml:"copyright"`
	PubDate     string    `xml:"pubDate"`
	Generator   string    `xml:"generator"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	Category    struct {
		Domain string `xml:"domain,attr"`
		Name   string `xml:",chardata"`
	} `xml:"category"`
	Comments  string `xml:"comments"`
	Enclosure []struct {
		URL  string `xml:"url,attr"`
		Len  int64  `xml:"length,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
	GUID struct {
		IsPermaLink bool   `xml:"type,attr"`
		Value       string `xml:",chardata"`
	} `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Source  string `xml:"source"`
	Torrent struct {
		Link          string `xml:"link"`
		PubDate       string `xml:"pubDate"`
		ContentLength string `xml:"contentLength"`
	} `xml:"torrent"`
}

var (
	timeFormat = []string{time.ANSIC, time.UnixDate, time.RubyDate,
		time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano,
		time.RFC822, time.RFC822Z, time.RFC850, time.Kitchen,
		"2006-01-02T15:04:05.999999999",
		time.Stamp, time.StampMicro, time.StampMilli, time.StampNano}
)

// ParseTime parses a string to time.Time.
// If fails to parse a string, it will return time.Now().
func ParseTime(s string) time.Time {
	for k := range timeFormat {
		t, e := time.Parse(timeFormat[k], s)
		if e == nil {
			return t
		}
	}
	return time.Now()
}
