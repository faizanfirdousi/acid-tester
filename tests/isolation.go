package tests

import (
	"database/sql"
	"fmt"

	"github.com/faizanfirdousi/acid-tester/color"
)

func TestIsolation(db *sql.DB) TestResult {
	color.Header("I — ISOLATION")
	color.Info("Concurrent transactions don't see each other's uncommitted changes.")

	db.Exec(`UPDATE bank_accounts SET balance = 8750.25 WHERE id = 4`)

	// ── READ COMMITTED ────────────────────────────────────────────────────────
	color.Step(1, 2, color.Bold("READ COMMITTED")+" — non-repeatable read expected")

	connA, err := db.Begin()
	if err != nil {
		return TestResult{Name: "Isolation", Passed: false, Details: fmt.Sprintf("Session A begin failed: %v", err)}
	}
	defer connA.Rollback()

	color.Query("BEGIN  -- Session A (READ COMMITTED, default)")
	color.Query("SELECT balance FROM bank_accounts WHERE id = 4  -- Session A, read #1")
	var balRead1 float64
	connA.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 4`).Scan(&balRead1)
	color.Info(fmt.Sprintf("Session A read #1 → Diana: %s", color.Yellow(fmt.Sprintf("$%.2f", balRead1))))

	color.Query("UPDATE bank_accounts SET balance = balance + 1000 WHERE id = 4  -- Session B commits")
	_, err = db.Exec(`UPDATE bank_accounts SET balance = balance + 1000 WHERE id = 4`)
	if err != nil {
		connA.Rollback()
		return TestResult{Name: "Isolation", Passed: false, Details: fmt.Sprintf("Session B failed: %v", err)}
	}
	color.Success("Session B committed +$1000 to Diana")

	color.Query("SELECT balance FROM bank_accounts WHERE id = 4  -- Session A, read #2")
	var balRead2 float64
	connA.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 4`).Scan(&balRead2)
	color.Info(fmt.Sprintf("Session A read #2 → Diana: %s", color.Yellow(fmt.Sprintf("$%.2f", balRead2))))

	if balRead2 != balRead1 {
		color.Warn(fmt.Sprintf("Non-repeatable read: $%.2f → $%.2f (expected under READ COMMITTED)", balRead1, balRead2))
	}
	connA.Rollback()

	// ── REPEATABLE READ ───────────────────────────────────────────────────────
	color.Step(2, 2, color.Bold("REPEATABLE READ")+" — snapshot must hold")

	db.Exec(`UPDATE bank_accounts SET balance = 8750.25 WHERE id = 4`)

	connA2, _ := db.Begin()
	defer connA2.Rollback()

	color.Query("BEGIN")
	color.Query("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")
	connA2.Exec(`SET TRANSACTION ISOLATION LEVEL REPEATABLE READ`)

	color.Query("SELECT balance FROM bank_accounts WHERE id = 4  -- Session A snapshot")
	var repBal1 float64
	connA2.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 4`).Scan(&repBal1)
	color.Info(fmt.Sprintf("Session A (RR) read #1 → Diana: %s", color.Yellow(fmt.Sprintf("$%.2f", repBal1))))

	color.Query("UPDATE bank_accounts SET balance = balance + 500 WHERE id = 4  -- Session B commits")
	db.Exec(`UPDATE bank_accounts SET balance = balance + 500 WHERE id = 4`)
	color.Success("Session B committed +$500 to Diana")

	color.Query("SELECT balance FROM bank_accounts WHERE id = 4  -- Session A, same snapshot")
	var repBal2 float64
	connA2.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 4`).Scan(&repBal2)
	color.Info(fmt.Sprintf("Session A (RR) read #2 → Diana: %s", color.Yellow(fmt.Sprintf("$%.2f", repBal2))))

	repeatableReadWorked := repBal1 == repBal2
	if repeatableReadWorked {
		color.Success(fmt.Sprintf("Snapshot preserved ($%.2f = $%.2f) — Session B invisible to A ✅", repBal1, repBal2))
	} else {
		color.Fail("Snapshot broken — REPEATABLE READ not working!")
	}
	connA2.Rollback()

	details := fmt.Sprintf(
		"READ COMMITTED: non-repeatable read ($%.2f → $%.2f). "+
			"REPEATABLE READ: snapshot held ($%.2f = $%.2f). Isolation %s.",
		balRead1, balRead2, repBal1, repBal2,
		map[bool]string{true: "WORKING ✅", false: "BROKEN ❌"}[repeatableReadWorked],
	)
	return TestResult{Name: "Isolation", Passed: repeatableReadWorked, Details: details}
}
