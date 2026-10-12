// Package search shows the core patterns behind a production hybrid search:
// keyword and vector retrieval fused with weighted Reciprocal Rank Fusion,
// typo tolerance that protects the first letter, a query-embedding cache with
// a lexical fallback, incremental re-embedding and offline nDCG evaluation.
//
// Rewritten from scratch to protect client code: names, schemas and
// infrastructure are replaced, and external services sit behind interfaces.
package search

import "sort"

// rrfK is the standard RRF damping constant: it keeps the top few ranks from
// dominating the fused score.
const rrfK = 60.0

// Weights tuned on the evaluation set: keyword 1.0 and semantic 0.5 beat
// equal weights, and keep exact text matches above purely semantic ones.
const (
	DefaultLexicalWeight  = 1.0
	DefaultSemanticWeight = 0.5
)

// Fuse merges a keyword ranking and a semantic ranking with weighted
// Reciprocal Rank Fusion: a result scores higher the earlier it appears in
// each list. Ties are broken by id so the order is deterministic.
func Fuse(lexical, semantic []string, wLex, wSem float64) []string {
	score := map[string]float64{}
	for i, id := range lexical {
		score[id] += wLex / (rrfK + float64(i+1))
	}
	for i, id := range semantic {
		score[id] += wSem / (rrfK + float64(i+1))
	}

	ids := make([]string, 0, len(score))
	for id := range score {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(a, b int) bool {
		if score[ids[a]] != score[ids[b]] {
			return score[ids[a]] > score[ids[b]]
		}
		return ids[a] < ids[b] // deterministic ties
	})
	return ids
}
