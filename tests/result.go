package tests

import (
	"fmt"

	"github.com/faizanfirdousi/acid-tester/color"
)

// TestResult holds metadata about one ACID test run.
type TestResult struct {
	Name    string
	Passed  bool
	Details string
}

// Print displays the result with color-coded pass/fail output.
func (r TestResult) Print() {
	if r.Passed {
		fmt.Printf("\n  %s  %s\n", color.Green("✔ PASSED"), color.Bold(r.Name))
	} else {
		fmt.Printf("\n  %s  %s\n", color.Red("✘ FAILED"), color.Bold(r.Name))
	}
	fmt.Printf("  %s %s\n", color.Gray("└─"), color.Gray(r.Details))
}
