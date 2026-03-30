-- =============================================================================
-- consistency.sql  —  ACID: Consistency
-- =============================================================================
-- WHAT IS CONSISTENCY?
--   The database always moves from one VALID state to another VALID state.
--   It enforces rules (constraints) so invalid data can NEVER exist.
--   Even if you try to insert garbage — Postgres will reject it.
--
-- TYPES OF CONSTRAINTS WE TEST:
--   CHECK       — enforces a condition on column values (e.g., balance >= 0)
--   NOT NULL    — a column must always have a value
--   FOREIGN KEY — a value must reference an existing row in another table
--   PRIMARY KEY — unique identifier, never null, never duplicated
--
-- Run manually:
--   psql -h localhost -U acid_user -d acid_db -f sql/consistency.sql
-- =============================================================================


-- ─────────────────────────────────────────────────────────────
-- VIOLATION 1: CHECK constraint — negative balance
-- ─────────────────────────────────────────────────────────────
-- Eve (id=5) has $620.00. Let's try to set her balance to -500.
-- The CHECK (balance >= 0) will block this.

-- Expected: ERROR: new row for relation "bank_accounts" violates check constraint
UPDATE bank_accounts
SET    balance = -500
WHERE  id = 5;

-- Verify Eve's balance is still $620.00 (unchanged)
SELECT name, balance FROM bank_accounts WHERE id = 5;


-- ─────────────────────────────────────────────────────────────
-- VIOLATION 2: NOT NULL constraint
-- ─────────────────────────────────────────────────────────────
-- Try inserting a customer without a name.
-- NOT NULL on the `name` column prevents this.

-- Expected: ERROR: null value in column "name" violates not-null constraint
INSERT INTO bank_accounts (name, balance)
VALUES (NULL, 1000.00);


-- ─────────────────────────────────────────────────────────────
-- VIOLATION 3: FOREIGN KEY constraint
-- ─────────────────────────────────────────────────────────────
-- Try creating a transaction that references account id=9999.
-- No such account exists — the foreign key makes this impossible.

-- Expected: ERROR: insert or update on table "transactions" violates
--           foreign key constraint "transactions_from_acc_fkey"
INSERT INTO transactions (from_acc, to_acc, amount, status)
VALUES (9999, 1, 50.00, 'success');


-- ─────────────────────────────────────────────────────────────
-- VIOLATION 4: CHECK constraint on transactions
-- ─────────────────────────────────────────────────────────────
-- Try inserting a transaction with a negative amount.
-- The CHECK (amount > 0) will reject it.

-- Expected: ERROR: new row for relation "transactions" violates check constraint
INSERT INTO transactions (from_acc, to_acc, amount, status)
VALUES (1, 2, -100.00, 'success');


-- ─────────────────────────────────────────────────────────────
-- LIST ALL CONSTRAINTS ON OUR TABLES (useful for learning!)
-- This is a meta-query — it reads Postgres's internal catalog.
-- ─────────────────────────────────────────────────────────────
SELECT
    tc.constraint_name,
    tc.table_name,
    tc.constraint_type,
    cc.check_clause
FROM
    information_schema.table_constraints AS tc
    LEFT JOIN information_schema.check_constraints AS cc
        ON tc.constraint_name = cc.constraint_name
WHERE
    tc.table_name IN ('bank_accounts', 'transactions')
ORDER BY
    tc.table_name, tc.constraint_type;

-- ─────────────────────────────────────────────────────────────
-- WHAT TO NOTICE:
--   Every single one of these inserts/updates was REJECTED.
--   The database is in the same valid state it was before.
--   Consistency = Postgres was the bouncer that blocked all bad data.
-- ─────────────────────────────────────────────────────────────
