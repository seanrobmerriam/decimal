# ADR 0004: Integer-Only Operations in Core

## Status
Accepted

## Date
2024-01-15

## Context
The Decimal Money Library is designed for financial calculations where precision is critical. The core `Money` type stores amounts as integers (in the smallest currency unit, typically cents), but we must decide whether operations like addition, subtraction, multiplication, and division should:

1. **Integer-only operations** (current approach): Only work on integer amounts
2. **Decimal operations**: Support fractional amounts directly

## Decision
The core `Money` type uses **integer-only storage and operations**. All amounts are stored as integers representing the smallest currency unit (cents, pence, yen, etc.).

### Rationale

1. **Floating-Point Precision Problems**

The infamous floating-point representation error:

```go
// JavaScript
0.1 + 0.2 // 0.30000000000000004

// Python
Decimal('0.1') + Decimal('0.2')  # Works, but float doesn't
0.1 + 0.2  # 0.30000000000000004

// Go
var f float64 = 0.1
fmt.Println(f + 0.2) // 0.30000000000000004
```

These errors accumulate in financial calculations:

```go
// Float: $100.00 - $0.01 - $0.01 - $0.01 ... 100 times
// Expected: $99.00, Actual: $98.99999999999999 or $99.00000000000001
```

2. **Binary to Decimal Conversion Errors**

Floating-point numbers are stored in binary (base 2), but humans think in decimal (base 10). Many decimal fractions cannot be represented exactly in binary:

| Fraction | Binary Representation | Rounded |
|----------|----------------------|---------|
| 0.1 | 0.000110011001100... | 0.00011001100110011... |
| 0.2 | 0.001100110011... | 0.0011001100110011... |
| 0.3 | 0.010011001100... | 0.01001100110011001... |

3. **Regulatory Compliance**

GAAP and IFRS require:
- Rounding to appropriate decimal places
- Consistent application of rounding methods
- Audit trails for all calculations

Integer-only math with explicit rounding satisfies all of these naturally.

4. **Performance**

Integer operations are:
- Faster than floating-point (no FPU needed for basic arithmetic)
- Deterministic across platforms
- Cache-friendly

## Consequences

### Positive
- No floating-point representation errors
- Exact decimal arithmetic
- Regulatory compliance by design
- Fast and deterministic
- Simple mental model (everything is cents)

### Negative
- Multiplication and division require explicit rounding
- Cannot represent fractions of a cent without scaling
- Need to handle overflow for very large amounts

### Mitigations
- `Multiply` and `Divide` methods accept rounding mode
- For crypto or stocks requiring fractional shares, use `MultiplyFloat` with caution
- Document overflow handling and limits
- Provide `BigMoney` for amounts exceeding int64 range

## Examples of Float Errors in Finance

### Example 1: The Classic 0.1 + 0.2 Problem

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    // What every programmer knows
    fmt.Printf("float64(0.1) = %.20f\n", float64(0.1))
    fmt.Printf("float64(0.2) = %.20f\n", float64(0.2))
    fmt.Printf("0.1 + 0.2 = %.20f\n", 0.1+0.2)
    fmt.Printf("0.1 + 0.2 == 0.3: %v\n", 0.1+0.2 == 0.3)
    
    // Financial calculation gone wrong
    price := 0.1
    tax := 0.2
    total := price + tax
    fmt.Printf("Price: %.2f, Tax: %.2f, Total: %.2f\n", price, tax, total)
    // Output: Price: 0.10, Tax: 0.20, Total: 0.30 (but internally 0.30000000000000004)
}
```

### Example 2: Accumulated Rounding Error

```go
package main

import (
    "fmt"
)

func main() {
    // Adding 0.01 one hundred times
    var floatSum float64
    for i := 0; i < 100; i++ {
        floatSum += 0.01
    }
    
    fmt.Printf("Float sum: %.20f\n", floatSum)
    fmt.Printf("Float sum == 1.0: %v\n", floatSum == 1.0)
    
    // Integer math: adding 1 cent one hundred times
    intSum := 0
    for i := 0; i < 100; i++ {
        intSum += 1 // 1 cent
    }
    
    fmt.Printf("Integer sum: %d cents = $%.2f\n", intSum, float64(intSum)/100)
    fmt.Printf("Integer sum == 100: %v\n", intSum == 100)
}
```

### Example 3: Financial Statement Error

```go
package main

import (
    "fmt"
)

func main() {
    // A company has 3 items at $0.33 each
    // Float calculation
    floatPrice := 0.33
    floatTotal := floatPrice * 3
    
    // Integer calculation (cents)
    intPrice := 33 // cents
    intTotal := intPrice * 3
    
    fmt.Printf("Float total: $%.2f (actual: %.20f)\n", floatTotal, floatTotal)
    fmt.Printf("Integer total: %d cents = $%.2f\n", intTotal, float64(intTotal)/100)
    
    // The float total should be $0.99, but...
    fmt.Printf("Float total == 0.99: %v\n", floatTotal == 0.99)
}
```

## Implementation Notes

### Money Type (Integer Storage)

```go
type Money struct {
    amount   int64  // Amount in smallest unit (cents)
    currency Currency
}

// New creates money from dollars and cents
func New(dollars int64, cents int64, currency string) (*Money, error) {
    // Convert to total cents
    totalCents := dollars*100 + cents
    // Handle negative amounts
    if dollars < 0 {
        totalCents = dollars*100 - cents
    }
    return &Money{
        amount:   totalCents,
        currency: NewCurrency(currency),
    }, nil
}
```

### Operations

```go
func (m *Money) Add(other *Money) (*Money, error) {
    if m.currency != other.currency {
        return nil, ErrCurrencyMismatch
    }
    return &Money{
        amount:   m.amount + other.amount,
        currency: m.currency,
    }, nil
}

func (m *Money) Multiply(decimal float64, mode RoundingMode) (*Money, error) {
    // Convert integer cents to decimal
    // Multiply
    // Round according to mode
    // Convert back to cents
}
```

## References

- [IEEE 754 Floating-Point Standard](https://ieeexplore.ieee.org/document/4610935)
- [What Every Computer Scientist Should Know About Floating-Point Arithmetic](https://docs.oracle.com/cd/E19957-01/800-3565/ncg_goldberg.html)
- [GAAP Rounding Guidelines](https://www.fasb.org/)
- [Java BigDecimal Documentation](https://docs.oracle.com/javase/8/docs/api/java/math/BigDecimal.html)
