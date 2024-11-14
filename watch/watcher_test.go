package watch

import (
	"context"
	"testing"
)

func TestWatcher(t *testing.T) {
	t.Log(WatchConfig(context.Background(), "test.txt", func() {
		t.Log("do")
	}))
}
