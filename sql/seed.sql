-- =============================================================================
-- seed.sql  —  Dummy data for testing
-- =============================================================================
-- "Seeding" = populating a database with initial/fake data for development/testing.
-- This is NOT a migration — it doesn't change schema, it just inserts rows.
--
-- Run after schema.sql:
--   psql -h localhost -U acid_user -d acid_db -f sql/seed.sql
-- =============================================================================


-- Insert 10 fake customers with varying balances.
-- Notice Charlie ($150) and Hank ($450) are low — useful for testing constraint violations.
-- Frank ($12000) is rich — used in Durability test.
INSERT INTO bank_accounts (name, balance) VALUES
    ('Alice',   5000.00),   -- id = 1
    ('Bob',     3200.50),   -- id = 2
    ('Charlie',  150.00),   -- id = 3  ← intentionally low for Atomicity test
    ('Diana',   8750.25),   -- id = 4  ← used in Isolation test
    ('Eve',      620.00),   -- id = 5  ← used in Consistency test
    ('Frank',  12000.00),   -- id = 6  ← used in Durability test
    ('Grace',    980.75),   -- id = 7  ← used in Durability test
    ('Hank',     450.00),   -- id = 8
    ('Ivy',     7300.00),   -- id = 9
    ('Jack',    2100.90);   -- id = 10


-- Insert some historical transaction records for analytics purposes.
-- These represent past transfers that already happened (status = 'success').
INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES
    (1, 2,   50.00, 'success'),
    (2, 3,  200.00, 'success'),
    (4, 5,   75.50, 'success'),
    (6, 7,  500.00, 'success'),
    (8, 9,   30.00, 'success'),
    (10, 1, 120.00, 'success'),
    (3, 6,   15.00, 'success'),
    (7, 2,  300.00, 'success'),
    (9, 4,   90.25, 'success'),
    (5, 10,  45.00, 'success');


-- Quick check — see what we seeded:
SELECT id, name, balance FROM bank_accounts ORDER BY id;
SELECT COUNT(*) AS total_transactions FROM transactions;
