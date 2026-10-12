package search

import (
	"context"
	"html"
	"log/slog"
	"regexp"
	"strings"
)

// Product is the part of a catalog item that feeds its embedding.
type Product struct {
	ID            string
	Name          string
	Description   string // may contain HTML
	Categories    []string
	EmbeddingHash string // hash of the text the stored vector was built from
}

// Catalog iterates products (PostgreSQL in production).
type Catalog interface {
	Each(ctx context.Context, fn func(Product) error) error
}

// VectorIndex stores a product's vector together with the hash it came from
// (Elasticsearch in production).
type VectorIndex interface {
	Update(ctx context.Context, id string, v []float32, hash string) error
}

// Stats is what one run did.
type Stats struct {
	Updated, Unchanged, Failed int
}

// Reembed is the hourly job that keeps product vectors fresh.
type Reembed struct {
	Catalog  Catalog
	Embedder Embedder
	Index    VectorIndex
	Log      *slog.Logger
}

// Run calls the embeddings API only for products whose text changed. One bad
// product never stops the run: it stays stale and is picked up next hour,
// because its stored hash still doesn't match.
func (j *Reembed) Run(ctx context.Context) (Stats, error) {
	var st Stats
	err := j.Catalog.Each(ctx, func(p Product) error {
		text := EmbeddingText(p)
		hash := sha256Hex(text)
		if hash == p.EmbeddingHash {
			st.Unchanged++
			return nil // no API call
		}

		v, err := j.Embedder.Embed(ctx, text)
		if err == nil {
			err = j.Index.Update(ctx, p.ID, v, hash)
		}
		if err != nil {
			st.Failed++
			j.Log.Warn("re-embed failed", "product", p.ID, "err", err)
			return nil
		}
		st.Updated++
		return nil
	})
	return st, err
}

var tags = regexp.MustCompile(`<[^>]*>`)

// EmbeddingText is the text a product's vector is built from. HTML is
// stripped first: before that, the tag word "strong" matched 73 products.
func EmbeddingText(p Product) string {
	desc := html.UnescapeString(tags.ReplaceAllString(p.Description, " "))
	parts := []string{p.Name, strings.Join(strings.Fields(desc), " "), strings.Join(p.Categories, ", ")}
	return strings.Join(parts, "\n")
}
