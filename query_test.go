package search

import (
	"encoding/json"
	"testing"
)

func TestFuzzyClauseProtectsTheFirstLetter(t *testing.T) {
	clauses := NameQuery("grey", []string{"name", "destination"})

	fuzzy := findClause(t, clauses, "fuzziness")
	if fuzzy["prefix_length"] != 1 {
		t.Fatalf("prefix_length = %v, want 1 so \"grey\" can't match \"rey\"", fuzzy["prefix_length"])
	}
	if fuzzy["fuzziness"] != "AUTO:4,8" {
		t.Fatalf("fuzziness = %v", fuzzy["fuzziness"])
	}
}

func TestExactMatchesRankAboveFuzzyOnes(t *testing.T) {
	clauses := NameQuery("grey", []string{"name"})

	exact := clauses[0]["multi_match"].(map[string]any)
	if exact["boost"] != 2 || exact["fuzziness"] != nil {
		t.Fatalf("first clause should be the boosted exact match, got %v", exact)
	}
}

func TestQueryIsValidJSON(t *testing.T) {
	if _, err := json.Marshal(map[string]any{
		"query": map[string]any{"bool": map[string]any{"should": NameQuery("tour de vinos", []string{"name"})}},
	}); err != nil {
		t.Fatal(err)
	}
}

func findClause(t *testing.T, clauses []map[string]any, key string) map[string]any {
	t.Helper()
	for _, c := range clauses {
		params := c["multi_match"].(map[string]any)
		if _, ok := params[key]; ok {
			return params
		}
	}
	t.Fatalf("no clause with %q", key)
	return nil
}
