-- =============================================================================
-- atomicity.sql  —  ACID: Atomicity
-- =============================================================================
-- WHAT IS ATOMICITY?
--   A transaction is ALL-OR-NOTHING.
--   If any step fails, ALL previous steps in the same transaction are undone.
--   The database either applies the ENTIRE transaction, or NONE of it.
--
-- KEY SQL CONCEPTS HERE:
--   BEGIN    — starts a transaction (nothing is saved to disk yet)
--   COMMIT   — makes all changes permanent
--   ROLLBACK — undoes everything since BEGIN (as if it never happened)
--
-- Run manually:
--   psql -h localhost -U acid_user -d acid_db -f sql/atomicity.sql
-- =============================================================================


-- ─────────────────────────────────────────────────────────────
-- TEST 1: SUCCESSFUL TRANSFER  (Alice → Bob, $200)
-- Both steps complete → COMMIT saves both changes.
-- ─────────────────────────────────────────────────────────────

-- Check balances BEFORE
SELECT name, balance FROM bank_accounts WHERE id IN (1, 2);
-- Expected: Alice=5000.00, Bob=3200.50


BEGIN;  -- Start transaction. Nothing written to DB until COMMIT.

    -- Step 1: Deduct $200 from Alice (id=1)
    UPDATE bank_accounts
    SET    balance = balance - 200
    WHERE  id = 1;

    -- Step 2: Add $200 to Bob (id=2)
    UPDATE bank_accounts
    SET    balance = balance + 200
    WHERE  id = 2;

    -- Log the transfer in audit table
    INSERT INTO transactions (from_acc, to_acc, amount, status)
    VALUES (1, 2, 200, 'success');

COMMIT;  -- Both updates are now permanently saved together ✅


-- Check balances AFTER
SELECT name, balance FROM bank_accounts WHERE id IN (1, 2);
-- Expected: Alice=4800.00, Bob=3400.50



-- ─────────────────────────────────────────────────────────────
-- TEST 2: FAILED TRANSFER  (Charlie → Bob, $99999)
-- Charlie doesn't have enough money.
-- The CHECK (balance >= 0) constraint will reject the debit.
-- ROLLBACK ensures Bob's balance does NOT change.
-- ─────────────────────────────────────────────────────────────

-- Check Charlie and Bob's balances BEFORE
SELECT name, balance FROM bank_accounts WHERE id IN (2, 3);
-- Expected: Charlie=150.00, Bob=3400.50 (from test 1 above)


BEGIN;  -- Start transaction

    -- Step 1: Try to deduct $99999 from Charlie (id=3)
    -- This WILL FAIL because 150 - 99999 = -99849, which violates CHECK (balance >= 0)
    -- Postgres will raise an error here.
    UPDATE bank_accounts
    SET    balance = balance - 99999
    WHERE  id = 3;

    -- Step 2: This line would credit Bob — but we NEVER reach it.
    -- Because the UPDATE above failed, Postgres already aborted this transaction.
    UPDATE bank_accounts
    SET    balance = balance + 99999
    WHERE  id = 2;

ROLLBACK;  -- Undo everything. Bob's balance is unchanged. ✅


-- Verify Bob's balance didn't change (atomicity working!)
SELECT name, balance FROM bank_accounts WHERE id IN (2, 3);
-- Expected: Charlie=150.00 (unchanged), Bob=3400.50 (unchanged — NOT 3400.50 + 99999)


-- ─────────────────────────────────────────────────────────────
-- WHAT TO NOTICE:
--   Even though step 1 in Test 2 ran before the error,
--   ROLLBACK reversed it completely.
--   Bob's account shows no sign of the attempted transfer.
--   This is ATOMICITY in action.
-- ─────────────────────────────────────────────────────────────
