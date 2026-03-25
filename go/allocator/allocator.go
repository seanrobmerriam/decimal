package allocator

import (
	"github.com/decimal/money/money"
)

// AllocateRatios distributes an amount according to given ratios.
// The ratios need not sum to any particular value.
//
// This function is useful for splitting expenses, distributing profits,
// or any scenario where money must be divided proportionally.
//
// The function uses the "largest remainder" method: each recipient
// receives floor(share), and the remaining pennies are distributed
// one at a time starting with the first recipient with the largest
// fractional remainder.
//
// Example: Splitting $100 among ratios 5:3:2
//
//	parts, _ := AllocateRatios(USD.FromString("100.00"), []int{5, 3, 2})
//	// parts[0] = $50.00, parts[1] = $30.00, parts[2] = $20.00
//
// Returns an error if ratios are empty, contain negative values,
// or sum to zero.
func AllocateRatios(amount money.Money, ratios []int) ([]money.Money, error) {
	if len(ratios) == 0 {
		return nil, money.InvalidOperationError{
			Operation: "allocate",
			Reason:    "ratios cannot be empty",
		}
	}

	// Calculate sum of ratios
	totalRatio := 0
	for _, r := range ratios {
		if r < 0 {
			return nil, money.InvalidOperationError{
				Operation: "allocate",
				Reason:    "ratios must be non-negative",
			}
		}
		totalRatio += r
	}

	if totalRatio == 0 {
		return nil, money.InvalidOperationError{
			Operation: "allocate",
			Reason:    "sum of ratios must be positive",
		}
	}

	result := make([]money.Money, len(ratios))
	remaining := amount.Amount()

	for i, r := range ratios {
		if i == len(ratios)-1 {
			// Last recipient gets whatever remains
			result[i] = money.Currency{}.FromInt(0) // This will be set properly
			if remaining != 0 {
				result[i] = amount.Currency().FromInt(remaining)
			} else {
				result[i] = amount.Currency().Zero()
			}
		} else {
			share := (amount.Amount() * int64(r)) / int64(totalRatio)
			result[i] = amount.Currency().FromInt(share)
			remaining -= share
			totalRatio -= r
		}
	}

	return result, nil
}

// Allocate evenly distributes an amount among n recipients.
// Any remainder is given to the first recipient.
//
// This function is useful when you need to split money into equal parts
// but must account for indivisible units (cents). The remainder is
// always given to the first recipient to ensure deterministic results.
//
// Example: Splitting $10 among 3 people
//
//	parts, _ := Allocate(USD.FromString("10.00"), 3)
//	// parts[0] = $3.34, parts[1] = $3.33, parts[2] = $3.33
//
// For proportional allocation with custom ratios, use AllocateRatios.
// Returns an error if n is not positive.
func Allocate(amount money.Money, n int) ([]money.Money, error) {
	if n <= 0 {
		return nil, money.InvalidOperationError{
			Operation: "allocate",
			Reason:    "n must be positive",
		}
	}

	if n == 1 {
		return []money.Money{amount}, nil
	}

	base := amount.Amount() / int64(n)
	remainder := amount.Amount() % int64(n)

	result := make([]money.Money, n)
	for i := 0; i < n; i++ {
		share := base
		if int64(i) < remainder {
			share++
		}
		result[i] = amount.Currency().FromInt(share)
	}

	return result, nil
}

// AllocateByPercentages distributes amount according to percentages.
// Percentages are expressed as integers (e.g., 50 for 50%).
//
// This function is useful for tax calculations, commission splits,
// or any scenario where allocations are defined as percentages.
//
// Percentages do not need to sum to 100; the function will allocate
// proportionally. The last recipient receives any remainder.
//
// Example: Distributing a $1000 bonus as percentages 40:30:30
//
//	parts, _ := AllocateByPercentages(USD.FromString("1000.00"), []int{40, 30, 30})
//	// parts[0] = $400.00, parts[1] = $300.00, parts[2] = $300.00
//
// Returns an error if percentages are empty, contain negative values,
// or sum to zero.
func AllocateByPercentages(amount money.Money, percentages []int) ([]money.Money, error) {
	if len(percentages) == 0 {
		return nil, money.InvalidOperationError{
			Operation: "allocate",
			Reason:    "percentages cannot be empty",
		}
	}

	// Calculate sum of percentages
	totalPercent := 0
	for _, p := range percentages {
		if p < 0 {
			return nil, money.InvalidOperationError{
				Operation: "allocate",
				Reason:    "percentages must be non-negative",
			}
		}
		totalPercent += p
	}

	if totalPercent == 0 {
		return nil, money.InvalidOperationError{
			Operation: "allocate",
			Reason:    "sum of percentages must be positive",
		}
	}

	result := make([]money.Money, len(percentages))
	remaining := amount.Amount()

	for i, p := range percentages {
		if i == len(percentages)-1 {
			// Last recipient gets whatever remains
			if remaining != 0 {
				result[i] = amount.Currency().FromInt(remaining)
			} else {
				result[i] = amount.Currency().Zero()
			}
		} else {
			share := (amount.Amount() * int64(p)) / int64(totalPercent)
			result[i] = amount.Currency().FromInt(share)
			remaining -= share
			totalPercent -= p
		}
	}

	return result, nil
}
