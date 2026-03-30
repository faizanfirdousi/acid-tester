package tests

import (
	"database/sql"
	"fmt"

	"github.com/faizanfirdousi/acid-tester/color"
)

func TestConsistency(db *sql.DB) TestResult {
	color.Header("C — CONSISTENCY")
	color.Info("DB constraints ensure the database only ever holds valid data.")

	allPassed := true

	// ── Violation 1: Negative balance ─────────────────────────────────────────
	color.Step(1, 3, "CHECK violation: set Eve's balance to "+color.Red("-500"))
	color.Query("UPDATE bank_accounts SET balance = -500 WHERE id = 5")
	_, err := db.Exec(`UPDATE bank_accounts SET balance = -500 WHERE id = 5`)
	if err != nil {
		color.Success("Blocked by CHECK (balance >= 0): " + color.Gray(shortenErr(err)))
	} else {
		color.Fail("BUG: negative balance allowed. Consistency BROKEN.")
		allPassed = false
	}

	// ── Violation 2: NULL name ────────────────────────────────────────────────
	color.Step(2, 3, "NOT NULL violation: insert account with "+color.Red("NULL name"))
	color.Query("INSERT INTO bank_accounts (name, balance) VALUES (NULL, 1000)")
	_, err = db.Exec(`INSERT INTO bank_accounts (name, balance) VALUES (NULL, 1000)`)
	if err != nil {
		color.Success("Blocked by NOT NULL constraint: " + color.Gray(shortenErr(err)))
	} else {
		color.Fail("BUG: NULL name allowed. Consistency BROKEN.")
		allPassed = false
	}

	// ── Violation 3: Bad foreign key ──────────────────────────────────────────
	color.Step(3, 3, "FOREIGN KEY violation: transaction referencing account "+color.Red("id=9999"))
	color.Query("INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (9999, 1, 50, 'success')")
	_, err = db.Exec(`INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES (9999, 1, 50, 'success')`)
	if err != nil {
		color.Success("Blocked by FOREIGN KEY constraint: " + color.Gray(shortenErr(err)))
	} else {
		color.Fail("BUG: orphan transaction allowed. Consistency BROKEN.")
		allPassed = false
	}

	color.Query("SELECT name, balance FROM bank_accounts WHERE id = 5")
	var eveBal float64
	db.QueryRow(`SELECT balance FROM bank_accounts WHERE id = 5`).Scan(&eveBal)
	color.Info(fmt.Sprintf("Eve's balance: %s %s", color.Yellow(fmt.Sprintf("$%.2f", eveBal)), color.Gray("← never touched")))

	details := fmt.Sprintf(
		"3 violations attempted: negative balance, NULL name, bad FK. "+
			"Eve's balance intact at $%.2f. All constraints %s.",
		eveBal,
		map[bool]string{true: "ENFORCED ✅", false: "BROKEN ❌"}[allPassed],
	)
	return TestResult{Name: "Consistency", Passed: allPassed, Details: details}
}

func shortenErr(err error) string {
	msg := err.Error()
	if len(msg) > 70 {
		return msg[:70] + "..."
	}
	return msg
}
