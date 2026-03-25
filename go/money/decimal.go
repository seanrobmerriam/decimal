package money

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Decimal represents an arbitrary-precision decimal number without currency association.
// Decimal is used for intermediate calculations, percentages, exchange rates, and
// other non-monetary values. Unlike Money, Decimal has no currency context.
//
// The Decimal type uses a "big.Int" unscaled value combined with a scale factor,
// allowing it to represent values like "1.23" as (123, 2) where 123/10^2 = 1.23.
// This approach provides arbitrary precision for intermediate calculations
// while the final result is typically converted to Money with currency context.
//
// Example:
//
//	rate := money.NewDecimalFromString("0.08875")        // 8.875%
//	factor := rate.Div(USD.FromInt(100), money.RoundHalfUp)
type Decimal struct {
	value *big.Int // Unscaled integer value (actual value = value / 10^scale)
	scale int32    // Number of decimal places
}

// NewDecimal creates a Decimal with the given unscaled value and scale.
func NewDecimal(value int64, scale int) *Decimal {
	return &Decimal{
		value: big.NewInt(value),
		scale: int32(scale),
	}
}

// NewDecimalFromBigInt creates a Decimal from a big.Int with the given scale.
func NewDecimalFromBigInt(value *big.Int, scale int) *Decimal {
	return &Decimal{
		value: new(big.Int).Set(value),
		scale: int32(scale),
	}
}

// NewDecimalFromString parses a string representation of a decimal number.
func NewDecimalFromString(s string) (*Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ParseError{
			Input:    s,
			Reason:   "empty string",
			Expected: "a decimal number",
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
		return nil, ParseError{
			Input:    s,
			Reason:   "invalid character in integer part",
			Expected: "digits only",
		}
	}

	// Validate and process fractional part
	scale := len(fracPart)
	if scale > 0 && !isValidDigits(fracPart) {
		return nil, ParseError{
			Input:    s,
			Position: len(intPart) + 1,
			Reason:   "invalid character in fractional part",
			Expected: "digits only",
		}
	}

	// Combine parts
	combined := intPart + fracPart
	val := new(big.Int)
	_, ok := val.SetString(combined, 10)
	if !ok {
		return nil, ParseError{
			Input:    s,
			Reason:   "invalid number format",
			Expected: "a valid decimal number",
		}
	}

	if negative {
		val.Neg(val)
	}

	return &Decimal{
		value: val,
		scale: int32(scale),
	}, nil
}

// NewDecimalFromFloat creates a Decimal from a float64 (WARNING: may lose precision).
func NewDecimalFromFloat(f float64) (*Decimal, error) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, ParseError{
			Input:    strconv.FormatFloat(f, 'f', -1, 64),
			Reason:   "cannot convert infinity or NaN to decimal",
			Expected: "a finite number",
		}
	}

	// Convert float to string with maximum precision
	s := strconv.FormatFloat(f, 'f', 15, 64)
	// Trim trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	return NewDecimalFromString(s)
}

// MustDecimalFromString parses a string or panics.
func MustDecimalFromString(s string) *Decimal {
	d, err := NewDecimalFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func isValidDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// String returns the decimal as a string representation.
func (d *Decimal) String() string {
	if d.scale == 0 {
		return d.value.String()
	}

	// Get the string representation
	s := d.value.String()
	if s == "0" && d.value.Sign() != 0 {
		// Handle negative zero case
		s = "-" + s
	}

	// If the number is negative and we're inserting a decimal point
	// we need to handle the sign separately
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	// Need to add leading zeros if necessary
	if len(s) <= int(d.scale) {
		// Pad with zeros at the front
		leadingZeros := int(d.scale) - len(s) + 1
		s = strings.Repeat("0", leadingZeros) + s
	}

	// Insert decimal point
	pointPos := len(s) - int(d.scale)
	result := s[:pointPos] + "." + s[pointPos:]

	if negative {
		result = "-" + result
	}

	return result
}

// Value returns the unscaled big.Int value.
func (d *Decimal) Value() *big.Int {
	return d.value
}

// Scale returns the number of decimal places.
func (d *Decimal) Scale() int32 {
	return d.scale
}

// IsZero returns true if the decimal value is zero.
func (d *Decimal) IsZero() bool {
	return d.value.Sign() == 0
}

// IsPositive returns true if the decimal value is greater than zero.
func (d *Decimal) IsPositive() bool {
	return d.value.Sign() > 0
}

// IsNegative returns true if the decimal value is less than zero.
func (d *Decimal) IsNegative() bool {
	return d.value.Sign() < 0
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d == 0
//	+1 if d > 0
func (d *Decimal) Sign() int {
	return d.value.Sign()
}

// Abs returns the absolute value of the decimal.
func (d *Decimal) Abs() *Decimal {
	return &Decimal{
		value: new(big.Int).Abs(d.value),
		scale: d.scale,
	}
}

// Neg returns the negated value of the decimal.
func (d *Decimal) Neg() *Decimal {
	return &Decimal{
		value: new(big.Int).Neg(d.value),
		scale: d.scale,
	}
}

// Add returns the sum of two decimals.
func (d *Decimal) Add(other *Decimal) *Decimal {
	// Align scales
	scale := d.scale
	if other.scale > scale {
		scale = other.scale
	}

	// Scale both values to the same precision
	scaleDiff1 := scale - d.scale
	scaleDiff2 := scale - other.scale

	val1 := new(big.Int).Set(d.value)
	val2 := new(big.Int).Set(other.value)

	if scaleDiff1 > 0 {
		val1.Mul(val1, big.NewInt(int64(math.Pow10(int(scaleDiff1)))))
	}
	if scaleDiff2 > 0 {
		val2.Mul(val2, big.NewInt(int64(math.Pow10(int(scaleDiff2)))))
	}

	result := new(big.Int).Add(val1, val2)
	return &Decimal{value: result, scale: scale}
}

// Sub returns the difference of two decimals.
func (d *Decimal) Sub(other *Decimal) *Decimal {
	return d.Add(other.Neg())
}

// Mul returns the product of two decimals.
func (d *Decimal) Mul(other *Decimal) *Decimal {
	result := new(big.Int).Mul(d.value, other.value)
	return &Decimal{
		value: result,
		scale: d.scale + other.scale,
	}
}

// Div returns the quotient of two decimals with the given rounding mode.
func (d *Decimal) Div(other *Decimal, rounding RoundingMode) (*Decimal, error) {
	if other.IsZero() {
		return nil, ParseError{
			Input:    "division",
			Reason:   "division by zero",
			Expected: "non-zero divisor",
		}
	}

	// Scale up the dividend to get more precision
	extraScale := int32(10) // Add extra precision for division

	// Multiply dividend by 10^extraScale before dividing
	scaled := new(big.Int).Mul(d.value, big.NewInt(int64(math.Pow10(int(extraScale)))))

	// Perform division
	quo := new(big.Int).Quo(scaled, other.value)

	// Round the result
	rounded := rounding.round(quo.Int64(), int(extraScale), 0)
	if rounded != quo.Int64() {
		quo = big.NewInt(rounded)
	}

	return &Decimal{value: quo, scale: d.scale}, nil
}

// MulInt multiplies the decimal by an integer.
func (d *Decimal) MulInt(factor int64) *Decimal {
	result := new(big.Int).Mul(d.value, big.NewInt(factor))
	return &Decimal{value: result, scale: d.scale}
}

// DivInt divides the decimal by an integer with the given rounding mode.
func (d *Decimal) DivInt(divisor int64, rounding RoundingMode) (*Decimal, error) {
	if divisor == 0 {
		return nil, ParseError{
			Input:    "division",
			Reason:   "division by zero",
			Expected: "non-zero divisor",
		}
	}

	// Simply divide the value
	quo := new(big.Int).Quo(d.value, big.NewInt(divisor))
	rem := new(big.Int).Rem(d.value, big.NewInt(divisor))

	// Apply rounding based on remainder
	if rem.Sign() != 0 {
		var needsInc bool
		absRem := new(big.Int).Abs(rem)
		absDiv := new(big.Int).Abs(big.NewInt(divisor))
		absRem.Mul(absRem, big.NewInt(2)) // For tie detection

		switch rounding {
		case RoundUp:
			needsInc = true
		case RoundDown:
			needsInc = false
		case RoundHalfUp:
			cmp := absRem.Cmp(absDiv)
			needsInc = cmp >= 0
		case RoundHalfDown:
			needsInc = absRem.Cmp(absDiv) > 0
		case RoundHalfEven:
			if absRem.Cmp(absDiv) > 0 {
				needsInc = true
			} else if absRem.Cmp(absDiv) == 0 && quo.Bit(0) == 1 { // odd quotient
				needsInc = true
			}
		case RoundCeiling:
			needsInc = d.value.Sign() > 0 && rem.Sign() != 0
		case RoundFloor:
			needsInc = d.value.Sign() < 0 && rem.Sign() != 0
		}

		if needsInc {
			if divisor > 0 {
				quo.Add(quo, big.NewInt(1))
			} else {
				quo.Sub(quo, big.NewInt(1))
			}
		}
	}

	return &Decimal{value: quo, scale: d.scale}, nil
}

// Cmp compares two decimals. Returns -1, 0, or 1.
func (d *Decimal) Cmp(other *Decimal) int {
	// Align scales
	scale := d.scale
	if other.scale > scale {
		scale = other.scale
	}

	// Scale both values
	scaleDiff1 := scale - d.scale
	scaleDiff2 := scale - other.scale

	val1 := new(big.Int).Set(d.value)
	val2 := new(big.Int).Set(other.value)

	if scaleDiff1 > 0 {
		val1.Mul(val1, big.NewInt(int64(math.Pow10(int(scaleDiff1)))))
	}
	if scaleDiff2 > 0 {
		val2.Mul(val2, big.NewInt(int64(math.Pow10(int(scaleDiff2)))))
	}

	return val1.Cmp(val2)
}

// Round rounds the decimal to the given number of decimal places.
func (d *Decimal) Round(rounding RoundingMode, decimals int) *Decimal {
	if decimals < 0 {
		decimals = 0
	}

	if int(d.scale) <= decimals {
		return d
	}

	rounded := rounding.round(d.value.Int64(), int(d.scale), decimals)
	return &Decimal{
		value: big.NewInt(rounded),
		scale: int32(decimals),
	}
}

// ToFloat converts the decimal to a float64 (may lose precision).
func (d *Decimal) ToFloat() float64 {
	f, _ := strconv.ParseFloat(d.String(), 64)
	return f
}

// ToInt64 returns the decimal as an int64, truncating the fractional part.
func (d *Decimal) ToInt64() int64 {
	if d.scale == 0 {
		return d.value.Int64()
	}

	// Truncate fractional part
	quo := new(big.Int).Quo(d.value, big.NewInt(int64(math.Pow10(int(d.scale)))))
	return quo.Int64()
}

// IntPart returns the integer part of the decimal.
func (d *Decimal) IntPart() int64 {
	if d.scale == 0 {
		return d.value.Int64()
	}
	// Truncate fractional part
	quo := new(big.Int).Quo(d.value, big.NewInt(int64(math.Pow10(int(d.scale)))))
	return quo.Int64()
}

// FracPart returns the fractional part as an int64 scaled appropriately.
func (d *Decimal) FracPart() int64 {
	if d.scale == 0 {
		return 0
	}

	// Get the fractional digits
	rem := new(big.Int).Rem(d.value, big.NewInt(int64(math.Pow10(int(d.scale)))))
	if rem.Sign() < 0 {
		rem.Neg(rem)
	}
	return rem.Int64()
}

// Pow10 returns 10^n as a new Decimal.
func Pow10(n int) *Decimal {
	return &Decimal{
		value: big.NewInt(int64(math.Pow10(n))),
		scale: 0,
	}
}
