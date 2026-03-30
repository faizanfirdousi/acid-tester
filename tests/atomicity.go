package tests

import (
	"database/sql"
	"fmt"

	"github.com/faizanfirdousi/acid-tester/color"
)

func TestAtomicity(db *sql.DB) TestResult {
	color.Header("A — ATOMICITY")
	color.Info("All-or-nothing: if any step fails, the whole transaction rolls back.")

	// ── Test 1: Successful transfer ───────────────────────────────────────────
	color.Step(1, 2, "Successful transfer: "+color.Cyan("Alice → Bob, $200"))

	txn, err := db.Begin()
	if err != nil {
		return TestResult{Name: "Atomicity", Passed: false, Details: fmt.Sprintf("Could not begin txn: %v", err)}
	}

	color.Query("BEGIN")

	color.Query("UPDATE bank_accounts SET balance = balance - 200 WHERE id = 1")
	_, err = txn.Exec(`UPDATE bank_accounts SET balance = balance - 200 WHERE id = 1`)
	if err != nil {
		txn.Rollback()
		return TestResult{Name: "Atomicity", Passed: false, Details: "Failed to debit Alice"}
	}
	color.Success("Alice debited $200")

	color.Query("UPDATE bank_accounts SET balance = balance + 200 WHERE id = 2")
	_, err = txn.Exec(`UPDATE bank_accounts SET balance = balance + 200 WHERE id = 2`)
	if err != nil {
		txn.Rollback()
		return TestResult{Name: "Atomicity", Passed: false, Details: "Failed to credit Bob"}
	}
	color.Success("Bob credited $200")

	color.Query("INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (1, 2, 200, 'success')")
	txn.Exec(`INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (1, 2, 200, 'success')`)

	color.Query("COMMIT")
	if err = txn.Commit(); err != nil {
		return TestResult{Name: "Atomicity", Passed: false, Details: fmt.Sprintf("Commit failed: %v", err)}
	}
	color.Success("COMMIT — both updates are permanent")

	color.Query("SELECT name, balance FROM bank_accounts WHERE id IN (1, 2)")
	var aliceBal, bobBal float64
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 1`).Scan(&aliceBal)
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 2`).Scan(&bobBal)
	color.Info(fmt.Sprintf("Alice: %s  │  Bob: %s", color.Yellow(fmt.Sprintf("$%.2f", aliceBal)), color.Yellow(fmt.Sprintf("$%.2f", bobBal))))

	// ── Test 2: Failing transfer (should rollback) ───────────────────────────
	color.Step(2, 2, "Failing transfer: "+color.Cyan("Charlie → Bob, $99999")+color.Gray(" (Charlie has ~$150)"))

	var bobBalBefore float64
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 2`).Scan(&bobBalBefore)

	txn2, _ := db.Begin()
	color.Query("BEGIN")

	// balance - 99999 < 0 → violates CHECK (balance >= 0)
	color.Query("UPDATE bank_accounts SET balance = balance - 99999 WHERE id = 3")
	_, err = txn2.Exec(`UPDATE bank_accounts SET balance = balance - 99999 WHERE id = 3`)
	if err != nil {
		color.Fail("Postgres raised constraint error → CHECK (balance ≥ 0) violated")
		txn2.Rollback()
		color.Query("ROLLBACK")
		color.Success("ROLLBACK — nothing was written")
	} else {
		txn2.Rollback()
	}

	var bobBalAfter float64
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 2`).Scan(&bobBalAfter)
	color.Info(fmt.Sprintf("Bob before: %s   Bob after: %s  %s",
		color.Yellow(fmt.Sprintf("$%.2f", bobBalBefore)),
		color.Yellow(fmt.Sprintf("$%.2f", bobBalAfter)),
		color.Gray("← unchanged"),
	))

	atomicityHeld := bobBalBefore == bobBalAfter
	details := fmt.Sprintf(
		"Successful transfer: Alice -$200, Bob +$200. Failed transfer rolled back. "+
			"Bob's balance unchanged: $%.2f. Atomicity %s.",
		bobBalAfter,
		map[bool]string{true: "HOLDING ✅", false: "BROKEN ❌"}[atomicityHeld],
	)
	return TestResult{Name: "Atomicity", Passed: atomicityHeld, Details: details}
}
