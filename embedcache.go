package search

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"
)

// Embedder turns text into a vector (OpenAI embeddings in production).
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Cache stores query vectors (Redis in production).
type Cache interface {
	Get(ctx context.Context, key string) ([]float32, bool)
	Set(ctx context.Context, key string, v []float32, ttl time.Duration)
}

const (
	embedTimeout  = 2 * time.Second
	queryCacheTTL = 7 * 24 * time.Hour
)

// Search holds what the query path needs.
type Search struct {
	Cache    Cache
	Embedder Embedder
	Log      *slog.Logger
}

// QueryVector returns the embedding for a search query, or false when it isn't
// available, in which case the caller serves keyword results only.
//
// The cache key is a hash of the normalized text, never the text itself, so
// raw searches never sit in the cache. An outage at the embedding provider
// makes ranking worse, but search never goes down.
func (s *Search) QueryVector(ctx context.Context, q string) ([]float32, bool) {
	key := "emb:" + sha256Hex(normalize(q))
	if v, ok := s.Cache.Get(ctx, key); ok {
		return v, true
	}

	embedCtx, cancel := context.WithTimeout(ctx, embedTimeout)
	defer cancel()
	v, err := s.Embedder.Embed(embedCtx, q)
	if err != nil {
		// Logged without the query; the caller goes lexical.
		s.Log.Warn("embedding unavailable", "err", err)
		return nil, false
	}

	s.Cache.Set(ctx, key, v, queryCacheTTL)
	return v, true
}

func normalize(q string) string {
	return strings.Join(strings.Fields(strings.ToLower(q)), " ")
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
