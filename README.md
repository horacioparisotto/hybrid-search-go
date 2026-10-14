# hybrid-search-go

[![CI](https://github.com/horacioparisotto/hybrid-search-go/actions/workflows/ci.yml/badge.svg)](https://github.com/horacioparisotto/hybrid-search-go/actions/workflows/ci.yml)

Patterns behind the hybrid semantic search I designed for **[Tur.com](https://www.tur.com/)**, a global, multi-tenant travel marketplace, on a Go backend with PostgreSQL, Elasticsearch and OpenAI embeddings.

> **Note.** Every sample here was written from scratch, independently, for this portfolio. None of it is client code, and it contains no client data, schemas, credentials or infrastructure details. The samples show general, widely used engineering patterns, simplified to illustrate my approach; the real implementations belong to the client and stay private. External services (Elasticsearch, Redis, the embeddings API) sit behind small interfaces with in-memory fakes, so everything here compiles and runs with no credentials.

**Results in production:** nDCG@5 0.41 → 0.75 · empty searches 31 → 0 · cached search ~35 ms.
Full case study: [horacioparisotto.com](https://www.horacioparisotto.com/)

```bash
go test ./...
```

## Patterns

### 1. Fusing keyword and semantic results · [`fusion.go`](fusion.go)

Hybrid search runs a keyword query and a vector query, then merges the two rankings with weighted **Reciprocal Rank Fusion**. A purely semantic neighbour must never outrank an exact text match, so the weights (keyword 1.0, semantic 0.5) were chosen on an evaluation set, not by eye, and the guarantee is a test. Ties are broken by id so results are deterministic.

### 2. Typo tolerance that keeps the first letter · [`query.go`](query.go)

Fuzzy matching allowed one edit in 4 to 7 letter words with no protected prefix, so the edit could land on the first letter: "grey" matched "rey". The fix requires the first letter to match and ranks exact words above fuzzy ones. The query lives in one shared helper that replaced four hand-written copies, so autocomplete and search results behave the same on the web and in the mobile app. Validated before release by replaying 3,662 real searches.

### 3. Query embeddings: cached, and never in the way · [`embedcache.go`](embedcache.go)

Query vectors are cached for 7 days under a hash of the normalized text, never the text itself. If the embeddings call fails or takes more than 2 s, search falls back to keyword results: an outage at the provider makes ranking worse, but search never goes down.

### 4. Re-embed only what changed · [`reembed.go`](reembed.go)

An hourly job hashes the text that feeds each embedding and calls the API only for products whose content changed. One failing product never stops the run; it keeps its old hash and is retried next hour. HTML is stripped first: before that, the tag word "strong" matched 73 products. A full catalog backfill cost about four cents, and a run with no changes makes no API calls.

### 5. The index nobody had added · [`migrations/0001_reviews_product_index.sql`](migrations/0001_reviews_product_index.sql)

Reviews had no index on the product they belong to, while every product page runs four queries filtered by it. Built with `CREATE INDEX CONCURRENTLY` so a live table keeps taking writes. Product-page review queries went from 5–20 ms to under 0.5 ms.

### 6. Measure ranking before shipping it · [`ndcg.go`](ndcg.go)

Every ranking change was scored offline before it shipped: 30 hand-written queries grouped by type, plus an independent holdout of 36 queries generated from randomly sampled products, graded 0–3 by an LLM judge checked against a stronger model (89% agreement, Cohen's κ = 0.77). 56 queries improved, 8 stayed the same and 2 got slightly worse, both only in the order of already-correct results.

## Layout

```
fusion.go        weighted Reciprocal Rank Fusion
query.go         Elasticsearch name query with a protected first letter
embedcache.go    query-embedding cache with a lexical fallback
reembed.go       incremental re-embedding job
ndcg.go          offline evaluation metric
migrations/      the missing index
*_test.go        tests for each pattern
```

---

Horacio Parisotto · Senior Full-Stack & AI Engineer · [Portfolio](https://www.horacioparisotto.com/) · [LinkedIn](https://www.linkedin.com/in/horacioparisotto/)
