// Package main provides a command-line calculator for monetary operations.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/decimal/money/allocator"
	"github.com/decimal/money/money"
)

// currentCurrency holds the active currency for the calculator session
var currentCurrency = money.USD

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║     Decimal Money Calculator           ║")
	fmt.Println("║     Precise Financial Arithmetic       ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <amount>          - Add two amounts")
	fmt.Println("  sub <amount>          - Subtract second from first")
	fmt.Println("  mul <amount> <factor> - Multiply by factor (decimal)")
	fmt.Println("  div <amount> <divisor> - Divide by divisor (decimal)")
	fmt.Println("  split <amount> <n>   - Split into n equal parts")
	fmt.Println("  alloc <amount> <r1,r2,...> - Allocate by ratios")
	fmt.Println("  set <currency>       - Set currency (USD, EUR, GBP, etc.)")
	fmt.Println("  show                 - Show current currency")
	fmt.Println("  help                 - Show this help")
	fmt.Println("  quit                 - Exit")
	fmt.Println()
	fmt.Println("Amounts are strings like: $10.50, 10.50, 10.5")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		var result string
		var err error

		switch cmd {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "help", "h", "?":
			showHelp()
			continue

		case "set":
			if len(args) < 1 {
				fmt.Println("Usage: set <currency>")
				continue
			}
			setCurrency(args[0])
			continue

		case "show":
			fmt.Printf("Current currency: %s (%s)\n", currentCurrency.Code(), currentCurrency.Name())
			continue

		case "add":
			result, err = handleAdd(args)
		case "sub":
			result, err = handleSub(args)
		case "mul":
			result, err = handleMul(args)
		case "div":
			result, err = handleDiv(args)
		case "split":
			result, err = handleSplit(args)
		case "alloc", "allocate":
			result, err = handleAlloc(args)
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
			continue
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Result: %s\n", result)
	}
}

func showHelp() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║           Calculator Commands           ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Monetary Operations:")
	fmt.Println("  add $10 $5.25       → $15.25")
	fmt.Println("  sub $10 $3.50       → $6.50")
	fmt.Println("  mul $100 0.08875    → $8.88 (with tax rate)")
	fmt.Println("  div $10 3           → $3.33")
	fmt.Println()
	fmt.Println("Allocation:")
	fmt.Println("  split $100 3        → $33.34, $33.33, $33.33")
	fmt.Println("  alloc $100 5,3,2    → $50.00, $30.00, $20.00")
	fmt.Println()
	fmt.Println("Currency:")
	fmt.Println("  set USD             → Use US Dollars")
	fmt.Println("  set EUR             → Use Euros")
	fmt.Println("  set JPY             → Use Japanese Yen (0 decimals)")
	fmt.Println()
}

func setCurrency(code string) {
	code = strings.ToUpper(code)

	var newCurrency money.Currency
	switch code {
	case "USD":
		newCurrency = money.USD
	case "EUR":
		newCurrency = money.EUR
	case "GBP":
		newCurrency = money.GBP
	case "JPY":
		newCurrency = money.JPY
	case "CHF":
		newCurrency = money.CHF
	case "CAD":
		newCurrency = money.CAD
	case "AUD":
		newCurrency = money.AUD
	case "CNY":
		newCurrency = money.CNY
	case "INR":
		newCurrency = money.INR
	case "BRL":
		newCurrency = money.BRL
	default:
		// Try to parse as custom currency
		parsed, err := money.ParseCurrency(code)
		if err != nil {
			fmt.Printf("Unknown currency: %s\n", code)
			fmt.Println("Available: USD, EUR, GBP, JPY, CHF, CAD, AUD, CNY, INR, BRL")
			return
		}
		newCurrency = parsed
	}

	currentCurrency = newCurrency
	fmt.Printf("Currency set to: %s (%s)\n", newCurrency.Code(), newCurrency.Name())
}

func parseMoney(s string) (money.Money, error) {
	// Remove currency symbols and whitespace
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "€", "")
	s = strings.ReplaceAll(s, "£", "")
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.TrimSpace(s)

	return currentCurrency.FromString(s)
}

func parseDecimal(s string) (*money.Decimal, error) {
	s = strings.TrimSpace(s)
	return money.NewDecimalFromString(s)
}

func handleAdd(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("add requires 2 amounts: add <a> <b>")
	}

	a, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid first amount: %v", err)
	}

	b, err := parseMoney(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid second amount: %v", err)
	}

	result, err := a.Add(b)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func handleSub(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("sub requires 2 amounts: sub <a> <b>")
	}

	a, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid first amount: %v", err)
	}

	b, err := parseMoney(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid second amount: %v", err)
	}

	result, err := a.Sub(b)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func handleMul(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("mul requires amount and factor: mul <amount> <factor>")
	}

	a, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid amount: %v", err)
	}

	factor, err := parseDecimal(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid factor: %v", err)
	}

	result, err := a.Mul(factor)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func handleDiv(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("div requires amount and divisor: div <amount> <divisor>")
	}

	a, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid amount: %v", err)
	}

	divisor, err := parseDecimal(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid divisor: %v", err)
	}

	if divisor.IsZero() {
		return "", fmt.Errorf("division by zero")
	}

	result, err := a.Div(divisor, money.RoundHalfUp)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func handleSplit(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("split requires amount and n: split <amount> <n>")
	}

	amount, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid amount: %v", err)
	}

	n, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid number: %v", err)
	}

	if n <= 0 {
		return "", fmt.Errorf("n must be positive")
	}

	parts, err := allocator.Allocate(amount, n)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, part := range parts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(part.String())
	}
	sb.WriteString("]")

	return sb.String(), nil
}

func handleAlloc(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("alloc requires amount and ratios: alloc <amount> <r1,r2,...>")
	}

	amount, err := parseMoney(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid amount: %v", err)
	}

	ratioStrings := strings.Split(args[1], ",")
	ratios := make([]int, len(ratioStrings))

	for i, rs := range ratioStrings {
		r, err := strconv.Atoi(strings.TrimSpace(rs))
		if err != nil {
			return "", fmt.Errorf("invalid ratio at position %d: %v", i+1, err)
		}
		ratios[i] = r
	}

	parts, err := allocator.AllocateRatios(amount, ratios)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i, part := range parts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(part.String())
	}
	sb.WriteString("]")

	return sb.String(), nil
}
