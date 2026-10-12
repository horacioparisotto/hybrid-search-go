package search

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
)

type sliceCatalog []Product

func (c sliceCatalog) Each(_ context.Context, fn func(Product) error) error {
	for _, p := range c {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

type memIndex struct {
	failFor string
	stored  map[string]string
}

func (m *memIndex) Update(_ context.Context, id string, _ []float32, hash string) error {
	if id == m.failFor {
		return errors.New("index unavailable")
	}
	m.stored[id] = hash
	return nil
}

func TestUnchangedProductsMakeNoAPICalls(t *testing.T) {
	p := Product{ID: "1", Name: "Wine tour", Description: "<p>Tasting</p>"}
	p.EmbeddingHash = sha256Hex(EmbeddingText(p))
	e := &fakeEmbedder{}

	st, err := newJob(sliceCatalog{p}, e, &memIndex{stored: map[string]string{}}).Run(context.Background())

	if err != nil || e.calls != 0 || st.Unchanged != 1 {
		t.Fatalf("stats=%+v calls=%d err=%v, want no API call", st, e.calls, err)
	}
}

func TestOneFailureDoesNotStopTheRun(t *testing.T) {
	idx := &memIndex{failFor: "2", stored: map[string]string{}}
	catalog := sliceCatalog{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}, {ID: "3", Name: "c"}}

	st, err := newJob(catalog, &fakeEmbedder{}, idx).Run(context.Background())

	if err != nil || st.Updated != 2 || st.Failed != 1 {
		t.Fatalf("stats=%+v err=%v, want 2 updated and 1 failed", st, err)
	}
	if _, ok := idx.stored["2"]; ok {
		t.Fatal("failed product must keep its old hash so the next run retries it")
	}
}

func TestEmbeddingTextStripsHTML(t *testing.T) {
	text := EmbeddingText(Product{Name: "Tour", Description: "<strong>Best</strong> view &amp; food"})

	if strings.Contains(text, "strong") || !strings.Contains(text, "Best view & food") {
		t.Fatalf("got %q", text)
	}
}

func newJob(c Catalog, e Embedder, idx VectorIndex) *Reembed {
	return &Reembed{Catalog: c, Embedder: e, Index: idx, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
}
