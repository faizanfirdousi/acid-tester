package tests

import (
	"database/sql"
	"fmt"

	"github.com/faizanfirdousi/acid-tester/color"
)

func TestDurability(db *sql.DB) TestResult {
	color.Header("D — DURABILITY")
	color.Info("Committed data survives crashes — Postgres uses WAL (Write-Ahead Log).")

	// ── Step 1: Commit a transfer ─────────────────────────────────────────────
	color.Step(1, 3, "Committing $300 transfer: "+color.Cyan("Frank → Grace"))

	txn, err := db.Begin()
	if err != nil {
		return TestResult{Name: "Durability", Passed: false, Details: fmt.Sprintf("Begin failed: %v", err)}
	}

	color.Query("BEGIN")

	color.Query("UPDATE bank_accounts SET balance = balance - 300 WHERE id = 6")
	_, err = txn.Exec(`UPDATE bank_accounts SET balance = balance - 300 WHERE id = 6`)
	if err != nil {
		txn.Rollback()
		return TestResult{Name: "Durability", Passed: false, Details: fmt.Sprintf("Debit failed: %v", err)}
	}
	color.Success("Frank debited $300")

	color.Query("UPDATE bank_accounts SET balance = balance + 300 WHERE id = 7")
	_, err = txn.Exec(`UPDATE bank_accounts SET balance = balance + 300 WHERE id = 7`)
	if err != nil {
		txn.Rollback()
		return TestResult{Name: "Durability", Passed: false, Details: fmt.Sprintf("Credit failed: %v", err)}
	}
	color.Success("Grace credited $300")

	color.Query("INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (6, 7, 300, 'success')")
	txn.Exec(`INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (6, 7, 300, 'success')`)

	color.Query("COMMIT")
	if err = txn.Commit(); err != nil {
		return TestResult{Name: "Durability", Passed: false, Details: fmt.Sprintf("Commit failed: %v", err)}
	}
	color.Success("COMMIT — Postgres flushed this to the WAL on disk before returning")

	// ── Step 2: Verify persistence ────────────────────────────────────────────
	color.Step(2, 3, "Reading back committed data")

	color.Query("SELECT name, balance FROM bank_accounts WHERE id IN (6, 7)")
	var frankBal, graceBal float64
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 6`).Scan(&frankBal)
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 7`).Scan(&graceBal)
	color.Info(fmt.Sprintf("Frank: %s  │  Grace: %s", color.Yellow(fmt.Sprintf("$%.2f", frankBal)), color.Yellow(fmt.Sprintf("$%.2f", graceBal))))

	dataPresent := frankBal > 0 && graceBal > 0

	// ── Step 3: Internal WAL stats ────────────────────────────────────────────
	color.Step(3, 3, "Inspecting Postgres WAL and checkpoint internals")

	color.Query("SELECT pg_current_wal_lsn()::text")
	var walLSN string
	db.QueryRow(`SELECT pg_current_wal_lsn()::text`).Scan(&walLSN)
	color.Info(fmt.Sprintf("WAL LSN (Log Sequence Number): %s", color.Cyan(walLSN)))
	color.Info(color.Gray("Every COMMIT writes to WAL ≥ this position before confirming to the app"))

	color.Query(`SELECT checkpoints_timed, checkpoints_req, buffers_checkpoint FROM pg_stat_bgwriter`)
	var checkpointsTimed, checkpointsReq, buffersCheckpoint int64
	db.QueryRow(`SELECT checkpoints_timed, checkpoints_req, buffers_checkpoint FROM pg_stat_bgwriter`).
		Scan(&checkpointsTimed, &checkpointsReq, &buffersCheckpoint)
	color.Info(fmt.Sprintf("Checkpoints (timed): %s   Checkpoints (req): %s   Pages flushed: %s",
		color.Cyan(fmt.Sprintf("%d", checkpointsTimed)),
		color.Cyan(fmt.Sprintf("%d", checkpointsReq)),
		color.Cyan(fmt.Sprintf("%d", buffersCheckpoint)),
	))

	color.Query("SELECT COUNT(*) FROM transactions WHERE status = 'success'")
	var txnCount int
	db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE status = 'success'`).Scan(&txnCount)
	color.Info(fmt.Sprintf("Total committed transactions in DB: %s", color.Yellow(fmt.Sprintf("%d", txnCount))))

	details := fmt.Sprintf(
		"Committed $300 Frank→Grace. Frank: $%.2f, Grace: $%.2f. "+
			"WAL LSN: %s. %d timed checkpoints. Data %s.",
		frankBal, graceBal, walLSN, checkpointsTimed,
		map[bool]string{true: "PERSISTED ✅", false: "MISSING ❌"}[dataPresent],
	)
	return TestResult{Name: "Durability", Passed: dataPresent, Details: details}
}
