package search

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

type memCache struct {
	mu   sync.Mutex
	data map[string][]float32
}

func newMemCache() *memCache { return &memCache{data: map[string][]float32{}} }

func (c *memCache) Get(_ context.Context, k string) ([]float32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[k]
	return v, ok
}

func (c *memCache) Set(_ context.Context, k string, v []float32, _ time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}

type fakeEmbedder struct {
	calls int
	err   error
	delay time.Duration
}

func (f *fakeEmbedder) Embed(ctx context.Context, _ string) ([]float32, error) {
	f.calls++
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	return []float32{0.1, 0.2}, nil
}

func newSearch(e *fakeEmbedder, c *memCache) *Search {
	return &Search{Cache: c, Embedder: e, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestSecondSearchIsServedFromCache(t *testing.T) {
	e := &fakeEmbedder{}
	s := newSearch(e, newMemCache())

	s.QueryVector(context.Background(), "Tour de Vinos")
	_, ok := s.QueryVector(context.Background(), "  tour de   vinos ")

	if !ok || e.calls != 1 {
		t.Fatalf("ok=%v calls=%d, want a cache hit for the same normalized query", ok, e.calls)
	}
}

func TestCacheNeverStoresTheRawQuery(t *testing.T) {
	c := newMemCache()
	newSearch(&fakeEmbedder{}, c).QueryVector(context.Background(), "private search text")

	for k := range c.data {
		if strings.Contains(k, "private") {
			t.Fatalf("cache key %q contains the query text", k)
		}
	}
}

func TestProviderErrorFallsBackToKeywordSearch(t *testing.T) {
	s := newSearch(&fakeEmbedder{err: errors.New("503")}, newMemCache())

	if _, ok := s.QueryVector(context.Background(), "wine tour"); ok {
		t.Fatal("want ok=false so the caller serves keyword results")
	}
}

func TestSlowProviderTimesOut(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the 2s timeout")
	}
	s := newSearch(&fakeEmbedder{delay: 10 * time.Second}, newMemCache())

	start := time.Now()
	_, ok := s.QueryVector(context.Background(), "wine tour")

	if ok || time.Since(start) > 3*time.Second {
		t.Fatalf("ok=%v after %v, want a fallback within the timeout", ok, time.Since(start))
	}
}
