// Package seed creates the database schema and inserts fake data.
package seed

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/faizanfirdousi/acid-tester/color"
)

// Run creates tables and stuffs them with dummy data.
// We use a bank scenario because it's the classic ACID example —
// money transfers touch multiple rows and must be all-or-nothing.
func Run(db *sql.DB) {
	fmt.Printf("%s Setting up schema and seeding dummy data...\n", color.Cyan("→"))

	// DROP + CREATE ensures a clean slate on every run
	schema := `
		DROP TABLE IF EXISTS transactions CASCADE;
		DROP TABLE IF EXISTS bank_accounts CASCADE;

		-- bank_accounts: each row is one customer's account
		-- CHECK (balance >= 0) is a DB-level constraint that enforces consistency
		CREATE TABLE bank_accounts (
			id      SERIAL PRIMARY KEY,
			name    TEXT NOT NULL,
			balance NUMERIC(12,2) NOT NULL CHECK (balance >= 0)
		);

		-- transactions: an audit log of every transfer attempt
		-- status can be 'success' or 'failed'
		CREATE TABLE transactions (
			id          SERIAL PRIMARY KEY,
			from_acc    INT  REFERENCES bank_accounts(id),
			to_acc      INT  REFERENCES bank_accounts(id),
			amount      NUMERIC(12,2) NOT NULL,
			status      TEXT NOT NULL DEFAULT 'pending',
			created_at  TIMESTAMP DEFAULT NOW()
		);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("❌ Schema creation failed: %v", err)
	}
	color.Success("Tables created: bank_accounts, transactions")

	// Seed 10 fake customer accounts with random balances
	customers := []struct {
		name    string
		balance float64
	}{
		{"Alice", 5000.00},
		{"Bob", 3200.50},
		{"Charlie", 150.00}, // intentionally low — useful for consistency tests
		{"Diana", 8750.25},
		{"Eve", 620.00},
		{"Frank", 12000.00},
		{"Grace", 980.75},
		{"Hank", 450.00},
		{"Ivy", 7300.00},
		{"Jack", 2100.90},
	}

	for _, c := range customers {
		_, err := db.Exec(
			`INSERT INTO bank_accounts (name, balance) VALUES ($1, $2)`,
			c.name, c.balance,
		)
		if err != nil {
			log.Fatalf("❌ Failed to insert customer %s: %v", c.name, err)
		}
	}
	color.Success(fmt.Sprintf("Inserted %d customer accounts", len(customers)))

	// Seed some historical transaction records
	// rand.Intn picks a random number — just for realistic fake data
	for i := 0; i < 20; i++ {
		from := rand.Intn(10) + 1
		to := rand.Intn(10) + 1
		for to == from {
			to = rand.Intn(10) + 1
		}
		amount := float64(rand.Intn(200)+1) + 0.50

		db.Exec(
			`INSERT INTO transactions (from_acc, to_acc, amount, status) VALUES ($1,$2,$3,'success')`,
			from, to, amount,
		)
	}
	color.Success("Inserted 20 historical transaction records")
	fmt.Println()
}
