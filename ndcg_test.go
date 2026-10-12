package search

import (
	"math"
	"testing"
)

func TestIdealRankingScoresOne(t *testing.T) {
	if got := NDCG([]float64{3, 2, 1, 0}, []float64{0, 1, 2, 3}, 5); !near(got, 1) {
		t.Fatalf("got %v, want 1", got)
	}
}

func TestMissingTheBestResultIsPenalized(t *testing.T) {
	// The only highly relevant document exists but wasn't returned.
	got := NDCG([]float64{1, 1, 0}, []float64{3, 1, 1, 0}, 5)
	if got >= 0.5 {
		t.Fatalf("got %v, want a low score when the best result is missing", got)
	}
}

func TestOrderMatters(t *testing.T) {
	good := NDCG([]float64{3, 0}, []float64{3, 0}, 5)
	bad := NDCG([]float64{0, 3}, []float64{3, 0}, 5)
	if !(good > bad) {
		t.Fatalf("good=%v bad=%v", good, bad)
	}
}

func TestNothingRelevantScoresZero(t *testing.T) {
	if got := NDCG([]float64{0, 0}, []float64{0, 0}, 5); got != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestMeanOverQueries(t *testing.T) {
	got := MeanNDCG([]Judged{
		{Ranked: []float64{3}, Judged: []float64{3}},
		{Ranked: []float64{0}, Judged: []float64{3}},
	}, 5)
	if !near(got, 0.5) {
		t.Fatalf("got %v, want 0.5", got)
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
