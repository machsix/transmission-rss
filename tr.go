package main

import (
	"context"
	"encoding/base64"
	"net/url"

	"github.com/hekmon/transmissionrpc/v3"
)

type Transmission struct {
	cli *transmissionrpc.Client
}

func NewTransmission(rpcUrl string) (*Transmission, error) {
	url, err := url.Parse(rpcUrl)
	if err != nil {
		return nil, err
	}

	cli, err := transmissionrpc.New(url, nil)
	if err != nil {
		return nil, err
	}

	return &Transmission{
		cli: cli,
	}, nil
}

func (t *Transmission) Add(ctx context.Context, files *Torrent, downloadDir string) error {
	str := base64.StdEncoding.EncodeToString(files.Bytes)
	_, err := t.cli.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
		DownloadDir: &downloadDir,
		MetaInfo:    &str,
	})

	return err
}
