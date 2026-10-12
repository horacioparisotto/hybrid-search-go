package search

// NameQuery builds the Elasticsearch "should" clauses used for names and
// destinations by both autocomplete and search results, so the two always
// behave the same. It replaced four hand-written copies of the same query.
//
// The bug it fixed: fuzzy matching allowed one edit in 4 to 7 letter words
// with no protected prefix, so the edit could land on the first letter and
// "grey" matched "rey". Requiring the first letter to match removes that,
// while exact words still rank above fuzzy ones.
func NameQuery(text string, fields []string) []map[string]any {
	return []map[string]any{
		// Exact words first.
		multiMatch(map[string]any{
			"query":  text,
			"fields": fields,
			"boost":  2,
		}),
		// Typo-tolerant words second.
		multiMatch(map[string]any{
			"query":         text,
			"fields":        fields,
			"fuzziness":     "AUTO:4,8", // 1 edit for 4–7 letters, 2 from 8
			"prefix_length": 1,          // the first letter must match
		}),
		// Words still being typed.
		multiMatch(map[string]any{
			"query":  text,
			"fields": fields,
			"type":   "phrase_prefix",
		}),
	}
}

func multiMatch(params map[string]any) map[string]any {
	return map[string]any{"multi_match": params}
}
