package search

import (
	"math"
	"sort"
)

// NDCG scores a ranking from 0 to 1: how close it is to the ideal order of
// everything judged for that query. ranked holds the grades (0–3) of the
// results in the order they were returned; judged holds the grades of every
// judged document for the query, returned or not.
func NDCG(ranked, judged []float64, k int) float64 {
	ideal := append([]float64(nil), judged...)
	sort.Sort(sort.Reverse(sort.Float64Slice(ideal)))
	best := dcg(ideal, k)
	if best == 0 {
		return 0
	}
	return dcg(ranked, k) / best
}

func dcg(grades []float64, k int) float64 {
	sum := 0.0
	for i := 0; i < k && i < len(grades); i++ {
		gain := math.Pow(2, grades[i]) - 1
		sum += gain / math.Log2(float64(i+2))
	}
	return sum
}

// MeanNDCG averages NDCG@k over a query set.
func MeanNDCG(queries []Judged, k int) float64 {
	if len(queries) == 0 {
		return 0
	}
	sum := 0.0
	for _, q := range queries {
		sum += NDCG(q.Ranked, q.Judged, k)
	}
	return sum / float64(len(queries))
}

// Judged is one evaluated query.
type Judged struct {
	Query  string
	Ranked []float64
	Judged []float64
}
