package money

import (
	"math"
)

// RoundingMode defines how to handle precision reduction when a monetary
// calculation produces a result with more decimal places than the currency
// supports. Different rounding modes are appropriate for different scenarios
// in financial applications.
type RoundingMode int

// Rounding mode constants. Each constant defines a specific rounding strategy.
const (
	// RoundUp rounds values away from zero. Also known as "round up" or "up".
	// Use when: You need to always round in the direction that increases magnitude.
	// Example: 1.001 -> 1.01, -1.001 -> -1.01
	// Common in: Tax calculations where partial cents always favor the government.
	RoundUp RoundingMode = iota

	// RoundDown rounds values toward zero. Also known as "truncate" or "floor".
	// Use when: You need to always reduce the magnitude, never rounding up.
	// Example: 1.019 -> 1.01, -1.019 -> -1.01
	// Common in: Discount calculations where you never want to overcharge.
	RoundDown

	// RoundHalfUp rounds values up when the discarded fraction is >= 0.5.
	// Use when: Standard rounding behavior is needed; this is the most intuitive mode.
	// Example: 1.015 -> 1.02, 1.014 -> 1.01
	// Common in: General-purpose rounding where "standard" behavior is expected.
	RoundHalfUp

	// RoundHalfDown rounds values up only when the discarded fraction is > 0.5.
	// Use when: You want ties to always round down (less common).
	// Example: 1.015 -> 1.01, 1.016 -> 1.02
	// Common in: Certain statistical applications.
	RoundHalfDown

	// RoundHalfEven rounds values to the nearest even number when exactly on a tie.
	// Also known as "banker's rounding" or "convergent rounding".
	// Use when: You need to minimize cumulative rounding bias over many operations.
	// Example: 1.015 -> 1.02 (5 rounds up), 1.025 -> 1.02 (2 rounds down)
	// Common in: Financial systems where many calculations average out; accounting.
	RoundHalfEven

	// RoundCeiling rounds values toward positive infinity.
	// Use when: You always want to round up for positive, round down for negative.
	// Example: 1.001 -> 1.01, -1.001 -> -1.00
	// Common in: Payment calculations where overpayment is better than underpayment.
	RoundCeiling

	// RoundFloor rounds values toward negative infinity.
	// Use when: You always want to round down for positive, round up for negative.
	// Example: 1.001 -> 1.00, -1.001 -> -1.01
	// Common in: Certain pricing strategies where you want to give the best rate.
	RoundFloor
)

// String returns a human-readable name for the rounding mode.
func (m RoundingMode) String() string {
	switch m {
	case RoundUp:
		return "RoundUp"
	case RoundDown:
		return "RoundDown"
	case RoundHalfUp:
		return "RoundHalfUp"
	case RoundHalfDown:
		return "RoundHalfDown"
	case RoundHalfEven:
		return "RoundHalfEven"
	case RoundCeiling:
		return "RoundCeiling"
	case RoundFloor:
		return "RoundFloor"
	default:
		return "Unknown"
	}
}

// round applies the rounding mode to produce a value with targetScale.
// The value is represented as (unscaledInt, scale) where the actual value is
// unscaledInt / (10^scale).
func (m RoundingMode) round(unscaledInt int64, scale, targetScale int) int64 {
	if scale <= targetScale {
		return unscaledInt
	}

	divisor := int64(math.Pow10(scale - targetScale))
	remainder := unscaledInt % divisor
	quotient := unscaledInt / divisor

	if m.needsIncrement(unscaledInt, divisor, remainder, quotient) {
		if unscaledInt < 0 {
			return quotient - 1
		}
		return quotient + 1
	}
	return quotient
}

func (m RoundingMode) needsIncrement(unscaled int64, divisor, remainder, quotient int64) bool {
	drop := remainder * 2 // For tie detection
	absQuotient := quotient

	switch m {
	case RoundUp:
		// RoundUp increments if remainder != 0 (always away from zero)
		return remainder != 0

	case RoundDown:
		return false

	case RoundHalfUp:
		return drop >= divisor

	case RoundHalfDown:
		return drop > divisor // Strictly greater than, not equal

	case RoundHalfEven:
		// Increment only if:
		// 1. remainder*2 > divisor (normal case), OR
		// 2. remainder*2 == divisor AND quotient is odd
		if drop > divisor {
			return true
		}
		if drop == divisor && absQuotient%2 != 0 {
			return true
		}
		return false

	case RoundCeiling:
		// Round toward +infinity: increment if positive and remainder != 0
		return unscaled >= 0 && remainder != 0

	case RoundFloor:
		// Round toward -infinity: increment if negative and remainder != 0
		return unscaled < 0 && remainder != 0
	}
	return false
}

// Round rounds the unscaled value to the given number of decimal places.
func (m RoundingMode) Round(unscaled int64, scale, targetScale int) int64 {
	return m.round(unscaled, scale, targetScale)
}
