# Go API Design Document

## Overview

This document describes the Go API design for the Decimal Money Library. The library provides precise decimal arithmetic for financial calculations, eliminating floating-point errors through integer-only internal operations.

## Design Principles

1. **Integer-only internal representation**: All monetary values are stored as integers with an associated scale (number of decimal places)
2. **Type safety**: Currency awareness at the type level to prevent mixing incompatible currencies
3. **Immutability**: All operations return new values; original values are never mutated
4. **Explicit over implicit**: Rounding modes, precision, and currency handling are explicit in API calls
5. **Performance**: Zero allocations in hot paths, efficient memory usage

---

## Type System

### Core Types

#### `Money`

```go
// Money represents a monetary amount with a specific currency.
// The internal representation uses an integer value and a scale (decimal places).
// Scale is determined by the currency (e.g., USD uses 2 decimal places).
type Money struct {
    amount   int64      // Unsigned internal representation, negative indicated by sign
    currency Currency   // Currency code (ISO 4217)
    // Private fields for internal state
    _       struct{}
}
```

**Design Decision**: Use `int64` for the amount field rather than `big.Int` for common operations.

- **Rationale**: `int64` covers most financial use cases (up to ~9 quintillion in smallest currency units)
- **Alternative Considered**: `big.Int` for arbitrary precision
  - Tradeoff: `big.Int` has allocation overhead on every operation
  - Decision: Use `int64` with overflow detection; users needing larger values can use `BigMoney`
- **Currency is part of the type**: This enforces compile-time safety against currency mixing

#### `BigMoney`

```go
// BigMoney provides arbitrary precision Money operations for extreme scale requirements.
// Use when amounts may exceed int64 range (e.g., national debt calculations).
type BigMoney struct {
    amount   *big.Int   // Arbitrary precision integer
    currency Currency
}
```

#### `Currency`

```go
// Currency represents an ISO 4217 currency code with associated properties.
type Currency struct {
    code     string     // ISO 4217 code (e.g., "USD", "EUR", "JPY")
    exponent int        // Number of decimal places (e.g., 2 for USD, 0 for JPY)
    name     string     // Full currency name
}

// Predefined currencies
var (
    USD = Currency{code: "USD", exponent: 2, name: "United States Dollar"}
    EUR = Currency{code: "EUR", exponent: 2, name: "Euro"}
    GBP = Currency{code: "GBP", exponent: 2, name: "British Pound"}
    JPY = Currency{code: "JPY", exponent: 0, name: "Japanese Yen"}
    CHF = Currency{code: "CHF", exponent: 2, name: "Swiss Franc"}
    CAD = Currency{code: "CAD", exponent: 2, name: "Canadian Dollar"}
    AUD = Currency{code: "AUD", exponent: 2, name: "Australian Dollar"}
    CNY = Currency{code: "CNY", exponent: 2, name: "Chinese Yuan"}
    INR = Currency{code: "INR", exponent: 2, name: "Indian Rupee"}
    BRL = Currency{code: "BRL", exponent: 2, name: "Brazilian Real"}
    // ... additional currencies as needed
)
```

#### `Decimal` (Currency-Agnostic)

```go
// Decimal represents an arbitrary-precision decimal number without currency association.
// Use for intermediate calculations or non-monetary decimal values.
type Decimal struct {
    value  *big.Int
    scale  int32  // Number of decimal places
}
```

---

## Package Structure

```
decimal/
├── money/           # Core Money type with currency awareness
│   ├── money.go     # Money type and basic operations
│   ├── currency.go  # Currency definitions and utilities
│   ├── ops.go       # Binary operations (add, sub, mul, div)
│   ├── rounding.go  # Rounding mode definitions
│   └── allocation.go # Distribution/splitting algorithms
├── decimal/         # Pure decimal without currency
│   ├── decimal.go   # Decimal type and operations
│   └── ops.go       # Decimal-specific operations
├── bigmoney/        # Arbitrary precision Money
│   ├── bigmoney.go
│   └── ops.go
├── conversion/      # Currency conversion
│   └── conversion.go
└── internal/
    ├── bigint/      # Optimized big integer operations
    └── testutil/    # Testing utilities
```

---

## Core Operations

### Construction

```go
// FromInt creates Money from an integer amount in the currency's smallest unit.
// Example: USD.FromInt(1999) = $19.99
func (c Currency) FromInt(amount int64) Money

// FromFloat creates Money from a float64 (approximate representation).
// WARNING: This may lose precision. Prefer FromInt for exact values.
func (c Currency) FromFloat(amount float64) (Money, error)

// FromString parses a string representation.
// Example: USD.FromString("19.99") = $19.99
func (c Currency) FromString(s string) (Money, error)

// FromDecimal creates Money from a Decimal, preserving currency.
func (c Currency) FromDecimal(d Decimal) Money

// Zero returns the zero value for a currency.
func (c Currency) Zero() Money

// ParseCurrency extracts currency code and returns Currency, error if unknown.
func ParseCurrency(code string) (Currency, error)
```

### Unary Operations

```go
// Neg returns the negated value.
func (m Money) Neg() Money

// Abs returns the absolute value.
func (m Money) Abs() Money

// IsZero returns true if amount is zero.
func (m Money) IsZero() bool

// IsPositive returns true if amount > 0.
func (m Money) IsPositive() bool

// IsNegative returns true if amount < 0.
func (m Money) IsNegative() bool
```

### Binary Operations

```go
// Add adds two Money values. Both must have the same currency.
func (m Money) Add(other Money) (Money, error)

// Sub subtracts other from m. Both must have the same currency.
func (m Money) Sub(other Money) (Money, error)

// Mul multiplies Money by a scalar (int, float, or Decimal).
func (m Money) Mul(scalar interface{}) (Money, error)

// Div divides Money by a scalar. Returns error if divisor is zero.
func (m Money) Div(scalar interface{}) (Money, error)

// Cmp compares two Money values. Returns -1, 0, or 1.
func (m Money) Cmp(other Money) int
```

**Error on Currency Mismatch**:

```go
// Attempting to add USD and EUR returns error
usd := USD.FromInt(1000)  // $10.00
eur := EUR.FromInt(1000)  // €10.00

result, err := usd.Add(eur)
if err != nil {
    // err: CurrencyMismatchError{Left: USD, Right: EUR, Operation: "add"}
}
```

### Allocation Operations

```go
// Split evenly distributes m into n parts, handling remainders.
// Remainder (cents) are distributed to the first parts.
func (m Money) Split(n int) []Money

// AllocateRatios distributes m according to given ratios (must sum to 100).
func (m Money) AllocateRatios(ratios []int) []Money

// AllocatePercentages distributes m according to percentages (must sum to 100%).
func (m Money) AllocatePercentages(percentages []Decimal) []Money

// Example: $10.07 split 3 ways = [$3.36, $3.36, $3.35]
money := USD.FromString("10.07")
parts := money.Split(3)
// parts[0] = $3.36
// parts[1] = $3.36
// parts[2] = $3.35
```

### Comparison Operations

```go
// Equal returns true if same currency and amount.
func (m Money) Equal(other Money) bool

// GreaterThan returns true if m > other.
func (m Money) GreaterThan(other Money) bool

// LessThan returns true if m < other.
func (m Money) LessThan(other Money) bool

// Compare returns: -1 (m < other), 0 (equal), 1 (m > other)
func (m Money) Compare(other Money) int
```

---

## Interface Design

### Interfaces for Extensibility

```go
// MoneyArithmetic defines the contract for money-like types.
type MoneyArithmetic interface {
    // Add returns the sum of two monetary values.
    Add(other MoneyArithmetic) (MoneyArithmetic, error)
    
    // Sub returns the difference.
    Sub(other MoneyArithmetic) (MoneyArithmetic, error)
    
    // Mul returns the product with a scalar.
    Mul(scalar interface{}) (MoneyArithmetic, error)
    
    // Div returns the quotient. Divisor of zero returns error.
    Div(divisor interface{}) (MoneyArithmetic, error)
    
    // Currency returns the currency of this value.
    Currency() Currency
    
    // IsZero returns true if the amount is zero.
    IsZero() bool
}

// ScalarArithmetic defines operations with scalar values (not Money).
type ScalarArithmetic interface {
    // AddInt64 adds an integer in the smallest currency unit.
    AddInt64(amount int64) (ScalarArithmetic, error)
    
    // MulInt64 multiplies by an integer.
    MulInt64(factor int64) (ScalarArithmetic, error)
    
    // DivInt64 divides by an integer, applying rounding.
    DivInt64(divisor int64, mode RoundingMode) (ScalarArithmetic, error)
}
```

---

## Rounding Modes

Rounding modes are specified explicitly in operations that may reduce precision:

```go
// RoundingMode defines how to handle precision reduction.
type RoundingMode int

const (
    // RoundUp (away from zero) - 1.001 -> 1.01, -1.001 -> -1.01
    RoundUp RoundingMode = iota
    // RoundDown (toward zero) - 1.019 -> 1.01, -1.019 -> -1.01
    RoundDown
    // RoundHalfUp - 1.015 -> 1.02, 1.014 -> 1.01
    RoundHalfUp
    // RoundHalfDown - 1.016 -> 1.01, 1.015 -> 1.01
    RoundHalfDown
    // RoundHalfEven (banker's rounding) - 1.015 -> 1.02, 1.025 -> 1.02
    RoundHalfEven
    // RoundCeiling (toward +∞) - 1.001 -> 1.01, -1.001 -> -1.00
    RoundCeiling
    // RoundFloor (toward -∞) - 1.001 -> 1.00, -1.001 -> -1.01
    RoundFloor
)

// WithRound applies rounding after an operation that may reduce precision.
func (m Money) WithRound(mode RoundingMode, scale int) Money
```

---

## Method Chaining

To support fluent API patterns:

```go
// MoneyBuilder allows chained operations.
type MoneyBuilder struct {
    money Money
}

// Pipeline example:
result := USD.FromInt(1000).
    Add(USD.FromInt(500)).
    Mul(Decimal.New(1.05)).
    Div(Decimal.New(4)).
    Round(RoundHalfUp, 2)
```

---

## Performance Considerations

### Zero-Allocation Hot Path

Critical operations avoid heap allocations:

```go
// Add is designed to be inlineable and allocation-free.
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{
        amount:   m.amount + other.amount,
        currency: m.currency,
    }, nil
}
```

### Inlining Guidelines

- Simple accessor methods (`Amount()`, `Currency()`, `IsZero()`) are simple enough for inlining
- Operations involving error checking may prevent inlining
- Use `go:inline` pragma for critical paths

### Memory Layout

```go
// Money is 16 bytes on 64-bit systems (single cache line friendly):
// - 8 bytes amount (int64)
// - 8 bytes currency pointer or inline representation
type Money struct {
    amount   int64
    currency Currency  // 8 bytes: 3 code bytes + 1 exp + padding
}
```

---

## String Representation

```go
// String returns human-readable format: "$19.99" for display.
func (m Money) String() string

// Format supports custom formatting.
// Example: m.Format(DecimalFormat{Symbol: "$", Thousands: true, Decimals: 2})
func (m Money) Format(fmt string) string

// MarshalText implements encoding.TextMarshaler for JSON serialization.
func (m Money) MarshalText() ([]byte, error)

// UnmarshalText implements encoding.TextUnmarshaler for JSON deserialization.
func (m Money) UnmarshalText(text []byte) error
```

---

## Usage Examples

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/decimal/money"
)

func main() {
    // Create money values
    price := money.USD.FromString("29.99")
    tax := money.USD.FromInt(225) // $2.25 (7.5% tax rate)
    shipping := money.USD.FromInt(599) // $5.99
    
    // Calculate total
    total, _ := price.Add(tax)
    total, _ = total.Add(shipping)
    
    fmt.Println(total) // "$38.23"
    
    // Apply 10% discount
    discount := total.Mul(money.Decimal.NewFloat(0.10))
    discounted, _ := total.Sub(discount.Round(money.RoundHalfUp, 2))
    
    fmt.Println(discounted) // "$34.41"
}
```

### Allocation Example

```go
// Split $100 among 3 people: [34, 33, 33]
amount := money.USD.FromString("100.00")
parts := amount.Split(3)

fmt.Printf("Person 1: %s\n", parts[0]) // $34.00
fmt.Printf("Person 2: %s\n", parts[1]) // $33.00
fmt.Printf("Person 3: %s\n", parts[2]) // $33.00
```

---

## Alternatives Considered

### Currency as Separate Parameter vs. Type Member

**Option A (Chosen)**: Currency as part of Money type
- Compile-time safety against mixing currencies
- Self-documenting code

**Option B**: Currency passed as separate parameter
- `Add(amount, currency)` 
- Runtime error only
- Rejected: Type safety is more important than API flexibility

### Amount as int64 vs. big.Int

**Chosen**: `int64` primary, `BigMoney` for overflow cases
- Better cache locality and performance for 99% of use cases
- Overflow detection with clear error
- `BigMoney` available for extreme cases

**Alternative**: Always use `big.Int`
- Simpler mental model
- Rejected: Performance overhead for common cases

---

## Summary

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Amount storage | `int64` | Performance, cache-friendly |
| Currency location | Type member | Compile-time safety |
| Immutability | Yes | Thread safety, predictability |
| Overflow handling | Error for int64, separate BigMoney type | Clear failure mode |
| Default rounding | `RoundHalfEven` | Industry standard for financial |
| Scale determination | From Currency | Reduces API complexity |