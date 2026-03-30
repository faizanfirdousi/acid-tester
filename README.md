# postgres-acid-tester 🧪🐘

A Go CLI that **actually runs and verifies** all four ACID properties of PostgreSQL — with real transactions, real constraints, and real internal database stats streamed to your terminal line by line.

Most people learn ACID from slides. This proves it works.

---

## Quickstart

**Prerequisites:** Go 1.22+, Docker

```bash
git clone https://github.com/faizanfirdousi/acid-tester
cd postgres-acid-tester

docker compose up -d   # start Postgres
go run .               # run the tests
```

That's it. No `.env` file, no config, no extra tools needed.

To stop when done:
```bash
docker compose down
```

---

## What it does

Seeds a fake bank database (10 accounts, 20 transactions), then runs four tests — one per ACID property. Every SQL query is printed in a styled box before it executes, and results stream in real-time.

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  A — ATOMICITY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [1/2] Successful transfer: Alice → Bob, $200

  ╭─ SQL ────────────────────────────────────────────────────────────────
  │  UPDATE bank_accounts SET balance = balance - 200 WHERE id = 1
  ╰──────────────────────────────────────────────────────────────────────
  ✔ Alice debited $200

  ╭─ SQL ──────
  │  COMMIT
  ╰────────────
  ✔ COMMIT — both updates are permanent

  ✔ PASSED  Atomicity
  └─ Successful transfer: Alice -$200, Bob +$200. Failed transfer rolled back.
```

---

## ACID tests

| Property | One-line summary | What we do |
|---|---|---|
| **A-Atomicity** | All-or-nothing | Transfer $200 Alice→Bob (succeeds). Transfer $99999 from Charlie with $150 (must rollback). Bob's balance must not change. |
| **C-Consistency** | Rules are always enforced | Attempt 3 violations: negative balance `CHECK`, `NULL` name, orphan foreign key. All must be rejected by Postgres. |
| **I-Isolation** | Concurrent sessions don't bleed | Session A reads Diana's balance. Session B commits a change. Under `READ COMMITTED` A sees it (expected). Under `REPEATABLE READ` A does not (snapshot preserved). |
| **D-Durability** | Committed data survives crashes | Commit Frank→Grace $300. Query `pg_current_wal_lsn()` and `pg_stat_bgwriter` to confirm WAL write position and checkpoint stats. |

---

## Stack

| | |
|---|---|
| Language | Go 1.22 |
| Database | PostgreSQL 16 (Docker) |
| DB driver | `lib/pq` — raw `database/sql`, no ORM |
| Terminal UI | Custom ANSI color package (zero external deps) |
| Config | Hardcoded defaults, env var overrides optional |

Only **one external dependency**: `github.com/lib/pq` (the Postgres driver).

---

## Project structure

```
postgres-acid-tester/
│
├── main.go                  # Entry point, test runner, analytics report
│
├── color/
│   └── color.go             # ANSI color helpers + SQL query box printer (zero deps)
│
├── db/
│   └── connect.go           # Builds DSN from defaults/env vars, returns *sql.DB
│
├── seed/
│   └── seed.go              # Drops + recreates tables, inserts 10 accounts + 20 txns
│
├── tests/
│   ├── result.go            # TestResult struct + colored PASSED/FAILED printer
│   ├── atomicity.go         # A: BEGIN/COMMIT + rollback on constraint failure
│   ├── consistency.go       # C: CHECK, NOT NULL, FOREIGN KEY violations
│   ├── isolation.go         # I: READ COMMITTED vs REPEATABLE READ
│   └── durability.go        # D: WAL LSN + pg_stat_bgwriter checkpoint stats
│
├── sql/                     # Plain SQL versions of every test (run in psql or any client)
│   ├── schema.sql           # CREATE TABLE — this is the actual migration file
│   ├── seed.sql             # INSERT dummy data
│   ├── atomicity.sql        # BEGIN / COMMIT / ROLLBACK walkthrough
│   ├── consistency.sql      # Constraint violation examples
│   ├── isolation.sql        # Two-session isolation demo with step-by-step instructions
│   ├── durability.sql       # WAL + checkpoint queries
│   └── analytics.sql        # Leaderboard, cache hit ratio, table sizes
│
├── docker-compose.yml       # PostgreSQL 16 with hardcoded local credentials
├── Makefile                 # Optional shortcuts (requires make to be installed)
├── go.mod / go.sum
└── .env.example             # Optional — only needed to override the defaults
```

---

## Run SQL files manually

All test logic also exists as plain `.sql` files — great for psql, TablePlus, DBeaver, or any SQL client.

```bash
# Open psql shell inside the container
docker exec -it acid_postgres psql -U acid_user -d acid_db

# Then run any file:
\i sql/atomicity.sql
\i sql/isolation.sql     # includes two-terminal instructions
\i sql/analytics.sql     # leaderboard, cache hit ratio, table sizes
```

Or pipe from outside:
```bash
docker exec -i acid_postgres psql -U acid_user -d acid_db < sql/atomicity.sql
```

## Override the database connection

The binary connects to `localhost:5432` with `acid_user / acid_pass / acid_db` by default — matching the Docker container exactly.

To use your own Postgres instance, set these environment variables before running:

```bash
DB_HOST=myhost DB_PORT=5432 DB_USER=me DB_PASSWORD=secret DB_NAME=mydb go run .
```

Or create a `.env` file (see `.env.example`).

---

## What you learn from this

- How `BEGIN` / `COMMIT` / `ROLLBACK` work at the SQL level
- How `CHECK`, `NOT NULL`, and `FOREIGN KEY` constraints enforce consistency
- The difference between `READ COMMITTED` and `REPEATABLE READ` isolation levels
- What PostgreSQL's Write-Ahead Log (WAL) is and how to read `pg_current_wal_lsn()`
- How to interpret `pg_stat_bgwriter` checkpoint statistics
- How Go's `database/sql` connection pool works with raw transactions

---

Built to understand databases deeply, not just theoretically.
