package search

import (
	"reflect"
	"testing"
)

func TestSemanticNeverOutranksExactMatch(t *testing.T) {
	lexical := []string{"tour-de-vinos"}
	semantic := []string{"wine-tasting", "vineyard-walk"}

	got := Fuse(lexical, semantic, DefaultLexicalWeight, DefaultSemanticWeight)

	if got[0] != "tour-de-vinos" {
		t.Fatalf("exact match ranked below %q", got[0])
	}
}

// With these weights, a document only the vector search found can't climb
// above anything in the first 60 keyword results.
func TestPurelySemanticResultsStayBelowTopKeywordResults(t *testing.T) {
	lexical := make([]string, 60)
	for i := range lexical {
		lexical[i] = string(rune('a'+i%26)) + string(rune('a'+i/26))
	}
	got := Fuse(lexical, []string{"semantic-only"}, DefaultLexicalWeight, DefaultSemanticWeight)

	if got[len(got)-1] != "semantic-only" {
		t.Fatalf("semantic-only result ranked at %d of %d", indexOf(got, "semantic-only"), len(got))
	}
}

func TestFoundByBothBeatsFoundByOne(t *testing.T) {
	got := Fuse([]string{"a", "both"}, []string{"both", "b"}, DefaultLexicalWeight, DefaultSemanticWeight)

	if got[0] != "both" {
		t.Fatalf("got %v, want the result both retrievers agree on first", got)
	}
}

func TestTiesAreDeterministic(t *testing.T) {
	first := Fuse([]string{"b"}, []string{"a"}, 1, 1)
	for i := 0; i < 50; i++ {
		if got := Fuse([]string{"b"}, []string{"a"}, 1, 1); !reflect.DeepEqual(got, first) {
			t.Fatalf("order changed between runs: %v vs %v", got, first)
		}
	}
	if !reflect.DeepEqual(first, []string{"a", "b"}) {
		t.Fatalf("got %v, want ties ordered by id", first)
	}
}

func indexOf(ids []string, id string) int {
	for i, v := range ids {
		if v == id {
			return i
		}
	}
	return -1
}
