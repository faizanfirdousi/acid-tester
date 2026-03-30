-- =============================================================================
-- analytics.sql  —  Post-run analytics + interesting DB stats
-- =============================================================================
-- Run this AFTER the tests to see a live snapshot of the database state.
-- These are all real queries against real Postgres internal tables.
--
-- Run manually:
--   psql -h localhost -U acid_user -d acid_db -f sql/analytics.sql
-- =============================================================================


-- ─────────────────────────────────────────────────────────────
-- ACCOUNT SUMMARY STATS
-- ─────────────────────────────────────────────────────────────

SELECT '=== ACCOUNT OVERVIEW ===' AS section;

SELECT
    COUNT(*)                         AS total_accounts,
    SUM(balance)                     AS total_money_in_system,
    ROUND(AVG(balance), 2)           AS average_balance,
    MAX(balance)                     AS highest_balance,
    MIN(balance)                     AS lowest_balance
FROM bank_accounts;


-- Leaderboard: richest to poorest
SELECT '=== BALANCE LEADERBOARD ===' AS section;

SELECT
    RANK() OVER (ORDER BY balance DESC) AS rank,
    name,
    balance,
    ROUND(balance / SUM(balance) OVER () * 100, 1) AS pct_of_total
FROM bank_accounts
ORDER BY balance DESC;


-- ─────────────────────────────────────────────────────────────
-- TRANSACTION STATS
-- ─────────────────────────────────────────────────────────────

SELECT '=== TRANSACTION STATS ===' AS section;

SELECT
    COUNT(*)                                            AS total_transactions,
    COUNT(*) FILTER (WHERE status = 'success')          AS successful,
    COUNT(*) FILTER (WHERE status = 'failed')           AS failed,
    COUNT(*) FILTER (WHERE status = 'pending')          AS pending,
    COALESCE(SUM(amount)  FILTER (WHERE status = 'success'), 0) AS total_volume,
    COALESCE(ROUND(AVG(amount), 2), 0)                  AS avg_transfer_amount,
    COALESCE(MAX(amount), 0)                            AS largest_transfer,
    COALESCE(MIN(amount), 0)                            AS smallest_transfer
FROM transactions;


-- Who sent the most money? (top senders)
SELECT '=== TOP SENDERS ===' AS section;

SELECT
    b.name         AS sender,
    COUNT(t.id)    AS transfers_made,
    SUM(t.amount)  AS total_sent
FROM transactions t
JOIN bank_accounts b ON b.id = t.from_acc
WHERE t.status = 'success'
GROUP BY b.name
ORDER BY total_sent DESC
LIMIT 5;


-- Who received the most? (top receivers)
SELECT '=== TOP RECEIVERS ===' AS section;

SELECT
    b.name         AS receiver,
    COUNT(t.id)    AS transfers_received,
    SUM(t.amount)  AS total_received
FROM transactions t
JOIN bank_accounts b ON b.id = t.to_acc
WHERE t.status = 'success'
GROUP BY b.name
ORDER BY total_received DESC
LIMIT 5;


-- ─────────────────────────────────────────────────────────────
-- POSTGRES INTERNAL HEALTH STATS
-- ─────────────────────────────────────────────────────────────

SELECT '=== POSTGRES INTERNALS ===' AS section;

-- Table bloat / live vs dead rows
-- "dead tuples" = rows marked for deletion but not yet vacuumed
SELECT
    relname          AS table_name,
    n_live_tup       AS live_rows,
    n_dead_tup       AS dead_rows,
    last_vacuum,
    last_autovacuum,
    last_analyze
FROM pg_stat_user_tables
WHERE relname IN ('bank_accounts', 'transactions');


-- Cache hit ratio (how often Postgres serves from RAM vs disk)
-- A healthy DB should have > 95% cache hit ratio
SELECT
    sum(heap_blks_read)                                    AS disk_reads,
    sum(heap_blks_hit)                                     AS cache_hits,
    ROUND(
        100.0 * sum(heap_blks_hit)
        / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0),
        2
    )                                                      AS cache_hit_ratio_pct
FROM pg_statio_user_tables
WHERE relname IN ('bank_accounts', 'transactions');
