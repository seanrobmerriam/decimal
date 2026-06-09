package money

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Money represents a monetary amount with a specific currency.
// The internal representation uses an integer value in the currency's smallest unit.
type Money struct {
	amount   int64    // Amount in smallest currency unit (cents for USD)
	currency Currency // ISO 4217 currency code and properties
}

// NewMoney creates a new Money value with the given amount in the currency's smallest unit.
func NewMoney(amount int64, currency Currency) (*Money, error) {
	return &Money{
		amount:   amount,
		currency: currency,
	}, nil
}

// FromInt creates Money from an integer amount in the currency's smallest unit.
// Example: USD.FromInt(1999) = $19.99
func (c Currency) FromInt(amount int64) Money {
	return Money{
		amount:   amount,
		currency: c,
	}
}

// FromFloat creates Money from a float64 (WARNING: may lose precision).
func (c Currency) FromFloat(amount float64) (Money, error) {
	if math.IsInf(amount, 0) || math.IsNaN(amount) {
		return Money{}, PrecisionLossError{
			Original:   strconv.FormatFloat(amount, 'f', -1, 64),
			Converted:  "0",
			LostDigits: -1,
			Context:    "FromFloat",
		}
	}

	// Convert float to string with maximum precision
	s := strconv.FormatFloat(amount, 'f', 15, 64)
	// Trim trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	return c.FromString(s)
}

// FromString parses a string representation.
// Example: USD.FromString("19.99") = $19.99
func (c Currency) FromString(s string) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Money{}, ParseError{
			Input:    s,
			Reason:   "empty string",
			Expected: "a monetary amount",
		}
	}

	// Handle negative
	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	// Split on decimal point
	parts := strings.Split(s, ".")
	intPart := parts[0]
	var fracPart string
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	// Validate integer part
	if intPart == "" {
		intPart = "0"
	}
	if !isValidDigits(intPart) {
		return Money{}, ParseError{
			Input:    s,
			Reason:   "invalid character in integer part",
			Expected: "digits only",
		}
	}

	// Validate and process fractional part
	if len(fracPart) > c.exponent {
		// Too many decimal places - we won't error but will truncate later with rounding
		fracPart = fracPart[:c.exponent]
	}
	if len(fracPart) > 0 && !isValidDigits(fracPart) {
		return Money{}, ParseError{
			Input:    s,
			Position: len(intPart) + 1,
			Reason:   "invalid character in fractional part",
			Expected: "digits only",
		}
	}

	// Pad fractional part with zeros to reach currency exponent
	for len(fracPart) < c.exponent {
		fracPart += "0"
	}

	// Combine parts
	combined := intPart + fracPart
	val, err := strconv.ParseInt(combined, 10, 64)
	if err != nil {
		return Money{}, ParseError{
			Input:    s,
			Reason:   "value out of range",
			Expected: "a valid monetary amount",
		}
	}

	if negative {
		val = -val
	}

	return Money{
		amount:   val,
		currency: c,
	}, nil
}

// FromDecimal creates Money from a Decimal, preserving currency.
func (c Currency) FromDecimal(d *Decimal) Money {
	return Money{
		amount:   d.ToInt64(),
		currency: c,
	}
}

// Zero returns the zero value for a currency.
func (c Currency) Zero() Money {
	return Money{
		amount:   0,
		currency: c,
	}
}

// Amount returns the amount in the currency's smallest unit.
func (m Money) Amount() int64 {
	return m.amount
}

// Currency returns the currency of this money value.
func (m Money) Currency() Currency {
	return m.currency
}

// String returns a string representation of the money value.
func (m Money) String() string {
	exp := m.currency.exponent
	absAmount := m.amount
	if absAmount < 0 {
		absAmount = -absAmount
	}

	var result string
	if exp == 0 {
		result = strconv.FormatInt(absAmount, 10)
	} else {
		// Format with decimal point
		amountStr := strconv.FormatInt(absAmount, 10)
		if len(amountStr) <= exp {
			// Need leading zeros
			amountStr = strings.Repeat("0", exp-len(amountStr)+1) + amountStr
		}

		// Insert decimal point
		pointPos := len(amountStr) - exp
		result = amountStr[:pointPos] + "." + amountStr[pointPos:]
	}

	if m.amount < 0 {
		result = "-" + result
	}

	return m.currency.code + " " + result
}

// Format returns a formatted string with symbol.
func (m Money) Format() string {
	exp := m.currency.exponent
	symbol := m.currency.code

	absAmount := m.amount
	if absAmount < 0 {
		absAmount = -absAmount
	}
	amountStr := strconv.FormatInt(absAmount, 10)
	if exp > 0 {
		if len(amountStr) <= exp {
			amountStr = strings.Repeat("0", exp-len(amountStr)+1) + amountStr
		}
		pointPos := len(amountStr) - exp
		amountStr = amountStr[:pointPos] + "." + amountStr[pointPos:]
	}

	if m.amount < 0 {
		amountStr = "-" + amountStr
	}

	return symbol + amountStr
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.amount == 0
}

// IsPositive returns true if the amount is greater than zero.
func (m Money) IsPositive() bool {
	return m.amount > 0
}

// IsNegative returns true if the amount is less than zero.
func (m Money) IsNegative() bool {
	return m.amount < 0
}

// Neg returns the negated value.
func (m Money) Neg() Money {
	return Money{
		amount:   -m.amount,
		currency: m.currency,
	}
}

// Abs returns the absolute value.
func (m Money) Abs() Money {
	if m.amount < 0 {
		return m.Neg()
	}
	return m
}

// Add adds two Money values. Both must have the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, CurrencyMismatchError{
			Left:      m.currency,
			Right:     other.currency,
			Operation: "add",
		}
	}

	// Check for overflow
	result := m.amount + other.amount
	if (result > m.amount) != (other.amount > 0 && m.amount > 0) &&
		(result < m.amount) != (other.amount < 0 && m.amount < 0) {
		// Overflow detected
		return Money{}, OverflowError{
			Operation: "add",
			Left:      m,
			Right:     &Decimal{value: big.NewInt(other.amount), scale: int32(m.currency.exponent)},
			MaxValue:  math.MaxInt64,
			Context: map[string]interface{}{
				"left_amount":  m.amount,
				"right_amount": other.amount,
			},
		}
	}

	return Money{
		amount:   result,
		currency: m.currency,
	}, nil
}

// Sub subtracts other from m. Both must have the same currency.
func (m Money) Sub(other Money) (Money, error) {
	return m.Add(other.Neg())
}

// Mul multiplies Money by a Decimal factor.
func (m Money) Mul(factor *Decimal) (Money, error) {
	// Scale up the money amount by the factor's scale
	factorScale := factor.scale

	// Multiply: m.amount * factor.value
	// Result scale = moneyScale + factorScale
	product := new(big.Int).Mul(big.NewInt(m.amount), factor.value)

	// The result needs to be rounded to the currency's scale
	// We have: product / 10^(moneyScale + factorScale)
	// We want: result / 10^moneyScale
	// So we need to divide by 10^factorScale and round

	if factorScale > 0 {
		divisor := big.NewInt(int64(math.Pow10(int(factorScale))))
		quo := new(big.Int).Quo(product, divisor)
		rem := new(big.Int).Rem(product, divisor)

		// Apply rounding
		rounded := RoundHalfEven.round(quo.Int64(), 0, 0)
		if rem.Sign() != 0 {
			// There's a remainder, need to consider rounding
			absRem := new(big.Int).Abs(rem)
			absRem.Mul(absRem, big.NewInt(2))
			absDiv := new(big.Int).Abs(divisor)
			if absRem.Cmp(absDiv) > 0 {
				rounded = quo.Int64() + 1
			} else if absRem.Cmp(absDiv) == 0 && quo.Bit(0) == 1 {
				rounded = quo.Int64() + 1
			}
		}
		return Money{
			amount:   rounded,
			currency: m.currency,
		}, nil
	}

	return Money{
		amount:   product.Int64(),
		currency: m.currency,
	}, nil
}

// MulInt multiplies Money by an integer factor.
func (m Money) MulInt(factor int64) (Money, error) {
	// Check for overflow
	if factor > 0 {
		if m.amount > math.MaxInt64/factor || m.amount < math.MinInt64/factor {
			return Money{}, OverflowError{
				Operation: "mul",
				Left:      m,
				Right:     &Decimal{value: big.NewInt(factor), scale: 0},
				MaxValue:  math.MaxInt64,
				Context: map[string]interface{}{
					"factor": factor,
				},
			}
		}
	} else if factor < 0 {
		// Handle negative factor carefully
		absFactor := -factor
		if m.amount > math.MaxInt64/absFactor || m.amount < math.MinInt64/absFactor {
			return Money{}, OverflowError{
				Operation: "mul",
				Left:      m,
				Right:     &Decimal{value: big.NewInt(factor), scale: 0},
				MaxValue:  math.MaxInt64,
				Context: map[string]interface{}{
					"factor": factor,
				},
			}
		}
	}

	return Money{
		amount:   m.amount * factor,
		currency: m.currency,
	}, nil
}

// Div divides Money by a divisor. Returns error if divisor is zero.
func (m Money) Div(divisor *Decimal, rounding RoundingMode) (Money, error) {
	if divisor.IsZero() {
		return Money{}, DivisionByZeroError{
			Dividend: m,
			Divisor:  divisor,
			Context:  "div",
		}
	}

	// Align scales: multiply amount by 10^divisor.scale
	pow10 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(divisor.scale)), nil)
	scaled := new(big.Int).Mul(big.NewInt(m.amount), pow10)

	// Perform division
	quo := new(big.Int).Quo(scaled, divisor.value)
	rem := new(big.Int).Rem(scaled, divisor.value)

	// Apply rounding based on remainder
	if rem.Sign() != 0 {
		var needsInc bool
		absRem := new(big.Int).Abs(rem)
		absRem.Mul(absRem, big.NewInt(2))
		absDiv := new(big.Int).Abs(divisor.value)

		switch rounding {
		case RoundUp:
			needsInc = true
		case RoundDown:
			needsInc = false
		case RoundHalfUp:
			needsInc = absRem.Cmp(absDiv) >= 0
		case RoundHalfDown:
			needsInc = absRem.Cmp(absDiv) > 0
		case RoundHalfEven:
			if absRem.Cmp(absDiv) > 0 {
				needsInc = true
			} else if absRem.Cmp(absDiv) == 0 && quo.Bit(0) == 1 {
				needsInc = true
			}
		case RoundCeiling:
			needsInc = quo.Sign() >= 0
		case RoundFloor:
			needsInc = quo.Sign() < 0
		}

		if needsInc {
			if quo.Sign() >= 0 {
				quo.Add(quo, big.NewInt(1))
			} else {
				quo.Sub(quo, big.NewInt(1))
			}
		}
	}

	return Money{
		amount:   quo.Int64(),
		currency: m.currency,
	}, nil
}

// DivInt divides Money by an integer divisor with rounding.
func (m Money) DivInt(divisor int64, rounding RoundingMode) (Money, error) {
	if divisor == 0 {
		return Money{}, DivisionByZeroError{
			Dividend: m,
			Divisor:  &Decimal{value: big.NewInt(0), scale: 0},
			Context:  "div",
		}
	}

	quo := m.amount / divisor
	rem := m.amount % divisor

	// Apply rounding based on remainder
	needsInc := false
	if rem != 0 {
		absDiv := divisor
		if divisor < 0 {
			absDiv = -divisor
		}
		absRem := rem
		if rem < 0 {
			absRem = -rem
		}

		// Compare absRem against absDiv/2 without overflow:
		// absRem*2 >= absDiv  ⇔  absRem >= absDiv - absRem
		absRemOpp := absDiv - absRem // always non-negative since absRem < absDiv

		switch rounding {
		case RoundUp:
			needsInc = true
		case RoundDown:
			needsInc = false
		case RoundHalfUp:
			needsInc = absRem >= absRemOpp
		case RoundHalfDown:
			needsInc = absRem > absRemOpp
		case RoundHalfEven:
			needsInc = absRem > absRemOpp || (absRem == absRemOpp && quo%2 != 0)
		case RoundCeiling:
			needsInc = quo >= 0
		case RoundFloor:
			needsInc = quo < 0
		}
	}

	if needsInc {
		if quo >= 0 {
			quo++
		} else {
			quo--
		}
	}

	return Money{
		amount:   quo,
		currency: m.currency,
	}, nil
}

// Cmp compares two Money values. Returns -1, 0, or 1.
func (m Money) Cmp(other Money) (int, error) {
	if m.currency != other.currency {
		return 0, CurrencyMismatchError{
			Left:      m.currency,
			Right:     other.currency,
			Operation: "compare",
		}
	}

	if m.amount < other.amount {
		return -1, nil
	}
	if m.amount > other.amount {
		return 1, nil
	}
	return 0, nil
}

// Equal returns true if the two Money values are equal.
func (m Money) Equal(other Money) bool {
	return m.currency == other.currency && m.amount == other.amount
}

// GreaterThan returns true if m > other.
func (m Money) GreaterThan(other Money) bool {
	cmp, _ := m.Cmp(other)
	return cmp > 0
}

// LessThan returns true if m < other.
func (m Money) LessThan(other Money) bool {
	cmp, _ := m.Cmp(other)
	return cmp < 0
}

// Split evenly distributes m into n parts, handling remainders.
// Remainder (cents) are distributed to the first parts.
func (m Money) Split(n int) ([]Money, error) {
	if n <= 0 {
		return nil, InvalidOperationError{
			Operation: "split",
			Reason:    "n must be positive",
		}
	}

	if n == 1 {
		return []Money{m}, nil
	}

	base := m.amount / int64(n)
	remainder := m.amount % int64(n)

	result := make([]Money, n)
	for i := 0; i < n; i++ {
		amount := base
		if remainder > 0 && int64(i) < remainder {
			amount++
		} else if remainder < 0 && int64(i) < -remainder {
			amount--
		}
		result[i] = Money{
			amount:   amount,
			currency: m.currency,
		}
	}

	return result, nil
}

// AllocateRatios distributes m according to given ratios (sum should be > 0).
func (m Money) AllocateRatios(ratios []int) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, InvalidOperationError{
			Operation: "allocate",
			Reason:    "ratios cannot be empty",
		}
	}

	// Calculate sum of ratios
	totalRatio := 0
	for _, r := range ratios {
		if r < 0 {
			return nil, InvalidOperationError{
				Operation: "allocate",
				Reason:    "ratios must be non-negative",
			}
		}
		totalRatio += r
	}

	if totalRatio == 0 {
		return nil, InvalidOperationError{
			Operation: "allocate",
			Reason:    "sum of ratios must be positive",
		}
	}

	result := make([]Money, len(ratios))
	remaining := m.amount

	for i, r := range ratios {
		if i == len(ratios)-1 {
			// Last recipient gets whatever remains
			result[i] = Money{
				amount:   remaining,
				currency: m.currency,
			}
		} else {
			share := (remaining * int64(r)) / int64(totalRatio)
			result[i] = Money{
				amount:   share,
				currency: m.currency,
			}
			remaining -= share
			totalRatio -= r
		}
	}

	return result, nil
}

// Round rounds the money to the currency's precision using the given mode.
func (m Money) Round(rounding RoundingMode) Money {
	// Money is already stored in minor units, so no rounding needed
	// unless we want to implement a custom scale
	return m
}

// ToDecimal converts Money to a Decimal representation.
func (m Money) ToDecimal() *Decimal {
	return &Decimal{
		value: big.NewInt(m.amount),
		scale: int32(m.currency.exponent),
	}
}
