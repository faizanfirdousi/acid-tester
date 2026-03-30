-- =============================================================================
-- schema.sql  —  This IS the "migration" file!
-- =============================================================================
-- A migration is a SQL file that defines or CHANGES your database schema.
-- This one creates the two tables our ACID tests rely on.
--
-- Run this manually with:
--   psql -h localhost -U acid_user -d acid_db -f sql/schema.sql
--
-- Or in psql shell:
--   \i sql/schema.sql
-- =============================================================================


-- Drop existing tables so we can start fresh.
-- CASCADE means: also drop anything that depends on these tables (like foreign keys).
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS bank_accounts CASCADE;


-- bank_accounts: represents customer bank accounts.
--
-- Columns explained:
--   id      — auto-incrementing primary key (SERIAL = integer that counts up)
--   name    — customer name, cannot be NULL (enforces data completeness)
--   balance — money amount with 2 decimal places (NUMERIC is exact, unlike FLOAT!)
--
-- CHECK (balance >= 0) is a CONSTRAINT — a rule Postgres enforces on every write.
-- If any INSERT or UPDATE would make balance negative, Postgres REJECTS it.
-- This is what makes the Consistency (C) in ACID possible at the DB level.
CREATE TABLE bank_accounts (
    id      SERIAL PRIMARY KEY,
    name    TEXT NOT NULL,
    balance NUMERIC(12, 2) NOT NULL CHECK (balance >= 0)
);


-- transactions: an audit trail of every money transfer attempt.
--
-- Columns explained:
--   id         — auto-incrementing primary key
--   from_acc   — which account sent money (REFERENCES = foreign key constraint)
--   to_acc     — which account received money
--   amount     — how much was transferred
--   status     — 'pending', 'success', or 'failed'
--   created_at — automatically set to current timestamp when row is inserted
--
-- REFERENCES bank_accounts(id) means:
--   You CANNOT insert a transaction that points to a non-existent account.
--   Postgres checks this automatically — another Consistency guard.
CREATE TABLE transactions (
    id         SERIAL PRIMARY KEY,
    from_acc   INT REFERENCES bank_accounts(id),
    to_acc     INT REFERENCES bank_accounts(id),
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    status     TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);


-- Verify the tables were created:
--   \dt           — list all tables
--   \d bank_accounts   — describe the bank_accounts table structure
