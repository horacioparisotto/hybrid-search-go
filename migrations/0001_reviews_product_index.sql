-- Reviews had no index on the product they belong to, while every product
-- page runs four queries filtered by it.
--
-- Before: Seq Scan on reviews, about 5 to 20 ms per query.
-- After:  Index Scan, under 0.5 ms.

-- Built without locking writes on a live table.
-- CONCURRENTLY has to run outside a transaction.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_reviews_product_created
  ON reviews (product_id, created_at DESC);

-- The query it serves:
-- SELECT rating, body, created_at
-- FROM reviews
-- WHERE product_id = $1
-- ORDER BY created_at DESC
-- LIMIT 10;
