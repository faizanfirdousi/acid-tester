-- =============================================================================
-- isolation.sql  —  ACID: Isolation
-- =============================================================================
-- WHAT IS ISOLATION?
--   Concurrent transactions don't interfere with each other.
--   Each transaction runs as if it's the ONLY one in the system.
--
-- ISOLATION LEVELS (weakest → strongest):
--   READ UNCOMMITTED — can see dirty (uncommitted) data from others  [Postgres ignores this]
--   READ COMMITTED   — only sees COMMITTED data                      [Postgres DEFAULT]
--   REPEATABLE READ  — same query always returns the same rows within a txn
--   SERIALIZABLE     — transactions appear to run one-at-a-time (strongest)
--
-- ANOMALIES these levels prevent:
--   ┌──────────────────────┬─────────┬──────────────┬────────────────┐
--   │ Anomaly              │ RC      │ RR           │ SERIALIZABLE   │
--   ├──────────────────────┼─────────┼──────────────┼────────────────┤
--   │ Dirty Read           │ prevent │ prevent      │ prevent        │
--   │ Non-Repeatable Read  │ ALLOWS  │ prevent      │ prevent        │
--   │ Phantom Read         │ ALLOWS  │ prevent      │ prevent        │
--   │ Serialization Anomaly│ ALLOWS  │ ALLOWS       │ prevent        │
--   └──────────────────────┴─────────┴──────────────┴────────────────┘
--
-- HOW TO READ THIS FILE:
--   To truly see isolation in action, open TWO psql terminal windows.
--   Run "Session A" lines in one window, "Session B" lines in the other.
--   The comments tell you what to run in each window.
--
-- Run manually (single window, educational):
--   psql -h localhost -U acid_user -d acid_db -f sql/isolation.sql
-- =============================================================================


-- ─────────────────────────────────────────────────────────────
-- SETUP: Reset Diana (id=4) to a known balance
-- ─────────────────────────────────────────────────────────────
UPDATE bank_accounts SET balance = 8750.25 WHERE id = 4;
SELECT name, balance FROM bank_accounts WHERE id = 4;  -- should be 8750.25


-- =============================================================
-- DEMO 1: READ COMMITTED (Non-Repeatable Read)
-- =============================================================
-- This is Postgres's DEFAULT isolation level.
-- A transaction can see NEW committed data mid-transaction.
-- This causes a "non-repeatable read" — the same SELECT returns different rows.
--
-- ┌─────────────────────────────────────────────────────────┐
-- │ SESSION A (paste into Terminal 1)                       │
-- └─────────────────────────────────────────────────────────┘

BEGIN;
-- SET TRANSACTION ISOLATION LEVEL READ COMMITTED;  ← default, no need to set
SELECT name, balance FROM bank_accounts WHERE id = 4;
-- [RESULT] Diana: 8750.25
-- ↑ Write this down. Now switch to Session B and run its block.
-- [After Session B commits, come back here:]
SELECT name, balance FROM bank_accounts WHERE id = 4;
-- [RESULT] Diana: 9750.25   ← changed! Non-repeatable read occurred.
-- This is EXPECTED under READ COMMITTED.
ROLLBACK;


-- ┌─────────────────────────────────────────────────────────┐
-- │ SESSION B  (paste into Terminal 2, while Session A waits)│
-- └─────────────────────────────────────────────────────────┘
-- (Run this AFTER Session A reads Diana's balance the first time)

UPDATE bank_accounts SET balance = balance + 1000 WHERE id = 4;
-- No BEGIN/COMMIT needed → auto-commits immediately
-- Diana's balance is now 9750.25, visible to everyone


-- ─────────────────────────────────────────────────────────────
-- SETUP: Reset Diana again for Demo 2
-- ─────────────────────────────────────────────────────────────
UPDATE bank_accounts SET balance = 8750.25 WHERE id = 4;


-- =============================================================
-- DEMO 2: REPEATABLE READ (Snapshot Isolation)
-- =============================================================
-- Session A takes a "snapshot" of the DB at BEGIN.
-- Even after Session B commits changes, Session A sees the OLD snapshot.
-- Same SELECT always returns the same result within this transaction.
--
-- ┌─────────────────────────────────────────────────────────┐
-- │ SESSION A (Terminal 1)                                  │
-- └─────────────────────────────────────────────────────────┘

BEGIN;
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

SELECT name, balance FROM bank_accounts WHERE id = 4;
-- [RESULT] Diana: 8750.25
-- ↑ Session A has locked in this snapshot. Now run Session B.
-- [After Session B commits:]
SELECT name, balance FROM bank_accounts WHERE id = 4;
-- [RESULT] Diana: 8750.25  ← SAME! Session B's commit is invisible here.
-- This is REPEATABLE READ working correctly ✅

ROLLBACK;  -- or COMMIT, doesn't matter since we only read


-- ┌─────────────────────────────────────────────────────────┐
-- │ SESSION B (Terminal 2, while Session A waits)           │
-- └─────────────────────────────────────────────────────────┘
-- (Run AFTER Session A reads Diana the first time in Demo 2)

UPDATE bank_accounts SET balance = balance + 500 WHERE id = 4;
-- Diana is now 9250.25 in the "real" DB — but Session A can't see it yet


-- =============================================================
-- INSPECT: Check current transaction isolation level
-- =============================================================
SHOW transaction_isolation;  -- shows current session's isolation level

-- See all active transactions (requires being superuser or pg_monitor role)
SELECT pid, query, state, wait_event_type, wait_event
FROM   pg_stat_activity
WHERE  state = 'active';

-- ─────────────────────────────────────────────────────────────
-- WHAT TO NOTICE:
--   Under READ COMMITTED: "Did my second SELECT return different data? YES"
--   Under REPEATABLE READ: "Did my second SELECT return different data? NO"
--   Isolation level controls how much of the outside world a transaction sees.
-- ─────────────────────────────────────────────────────────────
