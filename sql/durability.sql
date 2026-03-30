-- =============================================================================
-- durability.sql  —  ACID: Durability
-- =============================================================================
-- WHAT IS DURABILITY?
--   Once a transaction is COMMITTED, the data is permanently saved.
--   Even if the server crashes, power cuts out, or Docker restarts —
--   the committed data will still be there when you come back.
--
-- HOW POSTGRES ACHIEVES DURABILITY (WAL):
--   Postgres uses a Write-Ahead Log (WAL).
--
--   Step-by-step what happens on COMMIT:
--     1. Your changes are written to the WAL (a sequential log file on disk) FIRST.
--     2. Postgres replies "COMMIT successful" to your app.
--     3. Later (during a CHECKPOINT), the actual table data files are updated.
--
--   If the server crashes before step 3, it replays the WAL on restart.
--   The data is NEVER lost after step 1 completes.
--
-- Run manually:
--   psql -h localhost -U acid_user -d acid_db -f sql/durability.sql
-- =============================================================================


-- ─────────────────────────────────────────────────────────────
-- TEST: Commit a transfer and verify it persists
-- ─────────────────────────────────────────────────────────────

-- Check balances BEFORE
SELECT name, balance FROM bank_accounts WHERE id IN (6, 7);
-- Expected: Frank=12000.00, Grace=980.75


BEGIN;  -- Start a durable transaction

    -- Deduct $300 from Frank (id=6)
    UPDATE bank_accounts
    SET    balance = balance - 300
    WHERE  id = 6;

    -- Credit $300 to Grace (id=7)
    UPDATE bank_accounts
    SET    balance = balance + 300
    WHERE  id = 7;

    -- Log it
    INSERT INTO transactions (from_acc, to_acc, amount, status)
    VALUES (6, 7, 300.00, 'success');

COMMIT;
-- The moment COMMIT returns, Postgres guarantees this is on disk (in WAL).
-- Kill the server right now — this transfer will survive. ✅


-- Check balances AFTER
SELECT name, balance FROM bank_accounts WHERE id IN (6, 7);
-- Expected: Frank=11700.00, Grace=1280.75


-- ─────────────────────────────────────────────────────────────
-- INTERNAL POSTGRES DURABILITY STATS
-- These queries expose HOW Postgres manages durability internally.
-- ─────────────────────────────────────────────────────────────


-- 1. WAL: Write-Ahead Log current position
--    LSN = Log Sequence Number — a byte offset into the WAL file.
--    Every COMMIT writes to at least this LSN before confirming.
SELECT pg_current_wal_lsn() AS current_wal_lsn;


-- 2. WAL: How far has the WAL been flushed to disk?
--    lsn_diff > 0 means some data is buffered but not yet on disk.
SELECT
    pg_current_wal_lsn()   AS current_lsn,
    pg_current_wal_flush_lsn() AS flushed_lsn,
    (pg_current_wal_lsn() - pg_current_wal_flush_lsn()) AS lsn_diff_bytes;


-- 3. Checkpoint statistics from pg_stat_bgwriter
--    A CHECKPOINT is when Postgres flushes ALL dirty pages from RAM → disk.
--    checkpoints_timed  = scheduled checkpoints (every checkpoint_timeout seconds)
--    checkpoints_req    = forced checkpoints (WAL got too large)
--    buffers_checkpoint = number of 8KB pages written to disk at checkpoints
SELECT
    checkpoints_timed,
    checkpoints_req,
    buffers_checkpoint,
    buffers_clean,
    buffers_backend,
    pg_size_pretty(buffers_checkpoint * 8192::bigint) AS checkpoint_data_written
FROM pg_stat_bgwriter;


-- 4. WAL configuration — see how aggressive durability settings are
SELECT name, setting, unit, short_desc
FROM   pg_settings
WHERE  name IN (
    'synchronous_commit',   -- 'on' = wait for WAL flush before COMMIT returns
    'wal_level',            -- 'replica' or 'logical' = WAL verbosity
    'checkpoint_timeout',   -- how often automatic checkpoints run
    'max_wal_size',         -- max WAL size before forcing a checkpoint
    'fsync'                 -- 'on' = Postgres calls fsync() to harden data
)
ORDER BY name;


-- 5. Total amount of WAL ever generated
SELECT pg_size_pretty(pg_current_wal_lsn() - '0/0'::pg_lsn) AS total_wal_generated;


-- ─────────────────────────────────────────────────────────────
-- SIMULATE: What would happen on a crash?
-- (You can't do this in SQL, but here's what to do manually:)
--
--   1. Run the COMMIT block above
--   2. In terminal: docker compose restart postgres
--   3. Reconnect and run: SELECT balance FROM bank_accounts WHERE id IN (6,7)
--   4. Frank will be $11700, Grace $1280.75 — data survived the restart ✅
--
-- This proves Durability without needing to write any code.
-- ─────────────────────────────────────────────────────────────


-- 6. Useful: Check table sizes on disk
SELECT
    relname       AS table_name,
    pg_size_pretty(pg_total_relation_size(oid)) AS total_size,
    pg_size_pretty(pg_relation_size(oid))       AS table_size,
    pg_size_pretty(pg_indexes_size(oid))        AS index_size
FROM pg_class
WHERE relname IN ('bank_accounts', 'transactions')
ORDER BY pg_total_relation_size(oid) DESC;
