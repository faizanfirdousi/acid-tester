// Package color provides ANSI terminal color helpers with built-in streaming delays.
// No external dependencies — pure Go using standard escape codes + time.Sleep.
package color

import (
	"fmt"
	"strings"
	"time"
)

// Color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	white  = "\033[97m"
	gray   = "\033[90m"
)

// Delays — controls the streaming speed
const (
	delayHeader = 400 * time.Millisecond
	delayStep   = 250 * time.Millisecond
	delayLine   = 80 * time.Millisecond
	delayResult = 120 * time.Millisecond
	delayQuery  = 150 * time.Millisecond
)

// ── Color wrappers (return colored strings, no side effects) ──────────────────

func Bold(s string) string   { return bold + s + reset }
func Red(s string) string    { return red + s + reset }
func Green(s string) string  { return green + s + reset }
func Yellow(s string) string { return yellow + s + reset }
func Blue(s string) string   { return blue + s + reset }
func Cyan(s string) string   { return cyan + s + reset }
func Gray(s string) string   { return gray + s + reset }
func White(s string) string  { return white + s + reset }

// ── Print helpers (print AND pause for streaming effect) ─────────────────────

// Header prints a bold cyan section banner with a leading pause.
func Header(title string) {
	time.Sleep(delayHeader)
	line := cyan + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + reset
	fmt.Printf("\n%s\n  %s\n%s\n", line, bold+white+title+reset, line)
}

// Step prints a numbered step label.
func Step(n, total int, msg string) {
	time.Sleep(delayStep)
	fmt.Printf("\n  %s %s\n", Gray(fmt.Sprintf("[%d/%d]", n, total)), msg)
}

// Info prints an indented detail line.
func Info(msg string) {
	time.Sleep(delayLine)
	fmt.Printf("  %s %s\n", Gray("│"), msg)
}

// Success prints a green checkmark line.
func Success(msg string) {
	time.Sleep(delayResult)
	fmt.Printf("  %s %s\n", Green("✔"), msg)
}

// Warn prints a yellow warning line.
func Warn(msg string) {
	time.Sleep(delayResult)
	fmt.Printf("  %s %s\n", Yellow("⚠"), msg)
}

// Fail prints a red failure line.
func Fail(msg string) {
	time.Sleep(delayResult)
	fmt.Printf("  %s %s\n", Red("✘"), msg)
}

// Query prints a SQL statement in a styled box before it executes.
// Normalizes whitespace so multi-line SQL looks clean in the terminal.
//
// Example output:
//
//	╭─ SQL ──────────────────────────────────────────
//	│  UPDATE bank_accounts
//	│  SET balance = balance - 200 WHERE id = 1
//	╰────────────────────────────────────────────────
func Query(sql string) {
	time.Sleep(delayQuery)

	// Normalize: split on newlines, trim each line, drop blanks
	raw := strings.Split(sql, "\n")
	var lines []string
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return
	}

	// Find the longest line — box width adapts to content, no cap
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	bar := strings.Repeat("─", maxLen+4)
	fmt.Printf("  %s\n", blue+"╭─ SQL "+bar+reset)
	for _, l := range lines {
		fmt.Printf("  %s  %s\n", blue+"│"+reset, gray+l+reset)
	}
	fmt.Printf("  %s\n", blue+"╰"+bar+"──"+reset)
}
