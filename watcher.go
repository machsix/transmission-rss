package main

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchConfig(ctx context.Context, file string, do func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(file)
	if err != nil {
		return err
	}

	first := true

	timer := time.AfterFunc(-1, func() {
		if first {
			first = false
			return
		}

		do()
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			timer.Reset(time.Second * 5)
		}
	}
}
