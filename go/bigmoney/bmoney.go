package bigmoney

import (
	"math/big"
	"strings"

	decimalmoney "github.com/decimal/money/money"
)

// BigMoney provides arbitrary precision Money operations for extreme scale requirements.
// Use when amounts may exceed int64 range (e.g., national debt calculations,
// cryptocurrency with high precision, or large-scale financial projections).
//
// BigMoney uses the same mental model as Money but replaces int64 storage
// with big.Int, allowing for arbitrarily large (or small) monetary values.
// This comes at a performance cost, so use Money for normal financial amounts.
//
// Example:
//
//	// National debt calculation (trillions)
//	debt, _ := NewBigMoneyFromString(USD, "29300000000000.00")
//	interest, _ := debt.Mul(NewDecimalFromString("0.035"))
//
// Note: When the amount fits in int64 range, prefer Money for better performance.
type BigMoney struct {
	amount   *big.Int              // Arbitrary precision integer amount in smallest unit
	currency decimalmoney.Currency // ISO 4217 currency code and properties
}

// NewBigMoney creates a new BigMoney with the given amount and currency.
func NewBigMoney(amount *big.Int, currency decimalmoney.Currency) *BigMoney {
	return &BigMoney{
		amount:   new(big.Int).Set(amount),
		currency: currency,
	}
}

// NewBigMoneyFromString creates a BigMoney from a string representation.
func NewBigMoneyFromString(currency decimalmoney.Currency, s string) (*BigMoney, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, decimalmoney.ParseError{
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

	// Pad fractional part with zeros to reach currency exponent
	for len(fracPart) < currency.DecimalPlaces() {
		fracPart += "0"
	}
	// Truncate if too many
	if len(fracPart) > currency.DecimalPlaces() {
		fracPart = fracPart[:currency.DecimalPlaces()]
	}

	// Combine parts
	combined := intPart + fracPart
	val := new(big.Int)
	_, ok := val.SetString(combined, 10)
	if !ok {
		return nil, decimalmoney.ParseError{
			Input:    s,
			Reason:   "invalid number format",
			Expected: "a valid monetary amount",
		}
	}

	if negative {
		val.Neg(val)
	}

	return &BigMoney{
		amount:   val,
		currency: currency,
	}, nil
}

// Amount returns the amount as a big.Int.
func (b *BigMoney) Amount() *big.Int {
	return new(big.Int).Set(b.amount)
}

// Currency returns the currency.
func (b *BigMoney) Currency() decimalmoney.Currency {
	return b.currency
}

// IsZero returns true if the amount is zero.
func (b *BigMoney) IsZero() bool {
	return b.amount.Sign() == 0
}

// IsPositive returns true if the amount is greater than zero.
func (b *BigMoney) IsPositive() bool {
	return b.amount.Sign() > 0
}

// IsNegative returns true if the amount is less than zero.
func (b *BigMoney) IsNegative() bool {
	return b.amount.Sign() < 0
}

// String returns the string representation.
func (b *BigMoney) String() string {
	exp := b.currency.DecimalPlaces()
	if exp == 0 {
		return b.currency.Code() + " " + b.amount.String()
	}

	// Get string without sign
	abs := new(big.Int).Abs(b.amount)
	absStr := abs.String()
	if absStr == "0" {
		absStr = ""
	}

	// Format with decimal point
	if len(absStr) <= exp {
		absStr = strings.Repeat("0", exp-len(absStr)+1) + absStr
	}

	pointPos := len(absStr) - exp
	result := absStr[:pointPos] + "." + absStr[pointPos:]

	if b.amount.Sign() < 0 {
		result = "-" + result
	}

	return b.currency.Code() + " " + result
}

// Neg returns the negated value.
func (b *BigMoney) Neg() *BigMoney {
	return &BigMoney{
		amount:   new(big.Int).Neg(b.amount),
		currency: b.currency,
	}
}

// Abs returns the absolute value.
func (b *BigMoney) Abs() *BigMoney {
	return &BigMoney{
		amount:   new(big.Int).Abs(b.amount),
		currency: b.currency,
	}
}

// Add adds two BigMoney values.
func (b *BigMoney) Add(other *BigMoney) (*BigMoney, error) {
	if b.currency != other.currency {
		return nil, decimalmoney.CurrencyMismatchError{
			Left:      b.currency,
			Right:     other.currency,
			Operation: "add",
		}
	}

	result := new(big.Int).Add(b.amount, other.amount)
	return &BigMoney{
		amount:   result,
		currency: b.currency,
	}, nil
}

// Sub subtracts other from b.
func (b *BigMoney) Sub(other *BigMoney) (*BigMoney, error) {
	return b.Add(other.Neg())
}

// Mul multiplies by a factor.
func (b *BigMoney) Mul(factor *decimalmoney.Decimal) (*BigMoney, error) {
	// For BigMoney, we need to handle precision carefully
	result := new(big.Int).Mul(b.amount, factor.Value())
	// Adjust scale
	scale := factor.Scale()
	if scale > 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		result.Quo(result, divisor)
	}
	return &BigMoney{
		amount:   result,
		currency: b.currency,
	}, nil
}

// Div divides by a divisor with rounding.
func (b *BigMoney) Div(divisor *decimalmoney.Decimal, rounding decimalmoney.RoundingMode) (*BigMoney, error) {
	if divisor.IsZero() {
		return nil, decimalmoney.DivisionByZeroError{
			Dividend: decimalmoney.Money{},
			Divisor:  divisor,
			Context:  "bigmoney div",
		}
	}

	// Scale up dividend
	extraScale := int32(10)
	scaled := new(big.Int).Mul(b.amount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(extraScale)), nil))

	result := new(big.Int).Quo(scaled, divisor.Value())

	return &BigMoney{
		amount:   result,
		currency: b.currency,
	}, nil
}

// Cmp compares two BigMoney values.
func (b *BigMoney) Cmp(other *BigMoney) (int, error) {
	if b.currency != other.currency {
		return 0, decimalmoney.CurrencyMismatchError{
			Left:      b.currency,
			Right:     other.currency,
			Operation: "compare",
		}
	}
	return b.amount.Cmp(other.amount), nil
}

// ToMoney converts BigMoney to Money, potentially losing precision.
func (b *BigMoney) ToMoney() *decimalmoney.Money {
	m, _ := decimalmoney.NewMoney(b.amount.Int64(), b.currency)
	return m
}
