package main

import (
	"bytes"
	"encoding/gob"

	"go.etcd.io/bbolt"
)

type Cache interface {
	Load(rssUrl, torrentUrl string) (*Torrent, bool)
	Store(rssUrl, torrentUrl string, t *Torrent) error
	Close() error
}

type cache struct {
	b *bbolt.DB
}

func NewCacheByPath(path string) (Cache, error) {
	b, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &cache{b}, nil
}

func (c *cache) Load(rssUrl, torrentUrl string) (*Torrent, bool) {
	var t *Torrent
	_ = c.b.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte(rssUrl))
		if bkt == nil {
			return nil
		}

		data := bkt.Get([]byte(torrentUrl))
		if data == nil {
			return nil
		}

		tt := &Torrent{}

		err := gob.NewDecoder(bytes.NewReader(data)).Decode(tt)
		if err != nil {
			return err
		}

		t = tt
		return nil
	})

	return t, t != nil
}

func (c *cache) Store(rssUrl, torrentUrl string, t *Torrent) error {
	return c.b.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(rssUrl))
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		err = gob.NewEncoder(buf).Encode(t)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(torrentUrl), buf.Bytes())
	})
}

func (c *cache) Close() error {
	return c.b.Close()
}
