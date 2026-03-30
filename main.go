// acid-tester: A practical Go program that tests PostgreSQL's ACID properties.
//
// Run: go run .   (requires Docker container running first)
package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/faizanfirdousi/acid-tester/color"
	"github.com/faizanfirdousi/acid-tester/db"
	"github.com/faizanfirdousi/acid-tester/seed"
	"github.com/faizanfirdousi/acid-tester/tests"
)

func main() {
	printBanner()

	fmt.Printf("%s Connecting to PostgreSQL...\n", color.Cyan("→"))
	database := db.Connect()
	defer database.Close()
	fmt.Printf("%s Connected!\n", color.Green("✔"))

	seed.Run(database)

	start := time.Now()

	// Run each test and print its result IMMEDIATELY after it finishes.
	// This is what creates the live streaming feel — no batching.
	var results []tests.TestResult
	for _, fn := range []func(*sql.DB) tests.TestResult{
		tests.TestAtomicity,
		tests.TestConsistency,
		tests.TestIsolation,
		tests.TestDurability,
	} {
		r := fn(database) // runs the test — SQL and output stream in real-time
		r.Print()         // immediately prints PASSED/FAILED below each test
		results = append(results, r)
		time.Sleep(300 * time.Millisecond) // breath between tests
	}

	elapsed := time.Since(start)

	// Analytics
	printAnalytics(database)

	// Final score
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	fmt.Println()
	line := color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(line)
	if passed == len(results) {
		fmt.Printf("  %s  %s\n", color.Green(fmt.Sprintf("%d/%d ACID properties verified", passed, len(results))), color.Gray(fmt.Sprintf("(%s)", elapsed.Round(time.Millisecond))))
		fmt.Printf("  %s\n", color.Bold("PostgreSQL passes all ACID guarantees. 🐘"))
	} else {
		fmt.Printf("  %s  %s\n",
			color.Red(fmt.Sprintf("%d/%d passed", passed, len(results))),
			color.Gray(fmt.Sprintf("(%s)", elapsed.Round(time.Millisecond))))
	}
	fmt.Println(line)
	fmt.Println()
}

func printAnalytics(database *sql.DB) {
	color.Header("LIVE DB ANALYTICS")

	var accountCount int
	var totalBalance float64
	database.QueryRow(`SELECT COUNT(*), SUM(balance) FROM bank_accounts`).Scan(&accountCount, &totalBalance)
	color.Info(fmt.Sprintf("Accounts:             %s", color.Yellow(fmt.Sprintf("%d", accountCount))))
	color.Info(fmt.Sprintf("Total money in system:%s", color.Yellow(fmt.Sprintf(" $%.2f", totalBalance))))

	var richName string
	var richBal float64
	database.QueryRow(`SELECT name, balance FROM bank_accounts ORDER BY balance DESC LIMIT 1`).Scan(&richName, &richBal)
	color.Info(fmt.Sprintf("Richest account:      %s %s", color.Bold(richName), color.Gray(fmt.Sprintf("($%.2f)", richBal))))

	var poorName string
	var poorBal float64
	database.QueryRow(`SELECT name, balance FROM bank_accounts ORDER BY balance ASC LIMIT 1`).Scan(&poorName, &poorBal)
	color.Info(fmt.Sprintf("Lowest balance:       %s %s", color.Bold(poorName), color.Gray(fmt.Sprintf("($%.2f)", poorBal))))

	var txnTotal int
	var totalVolume float64
	database.QueryRow(`SELECT COUNT(*), COALESCE(SUM(amount),0) FROM transactions WHERE status='success'`).Scan(&txnTotal, &totalVolume)
	color.Info(fmt.Sprintf("Committed transfers:  %s %s", color.Yellow(fmt.Sprintf("%d", txnTotal)), color.Gray(fmt.Sprintf("($%.2f total volume)", totalVolume))))

	var walLSN string
	database.QueryRow(`SELECT pg_current_wal_lsn()::text`).Scan(&walLSN)
	color.Info(fmt.Sprintf("WAL position (LSN):   %s", color.Cyan(walLSN)))

	var pgVersion string
	database.QueryRow(`SELECT split_part(version(), ' ', 2)`).Scan(&pgVersion)
	color.Info(fmt.Sprintf("PostgreSQL version:   %s", color.Cyan(pgVersion)))
}

func printBanner() {
	cyan := "\033[36m"
	bold := "\033[1m"
	reset := "\033[0m"
	gray := "\033[90m"

	fmt.Print(cyan + bold + `
  ╔══════════════════════════════════════════════════╗
  ║        🧪  POSTGRES ACID TESTER  🐘             ║
  ╚══════════════════════════════════════════════════╝` + reset + "\n")
	fmt.Println(gray + "  Atomicity · Consistency · Isolation · Durability" + reset)
	fmt.Println()
}
