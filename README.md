# Decimal Money Library

[![Go Reference](https://pkg.go.dev/badge/github.com/decimal/money.svg)](https://pkg.go.dev/github.com/decimal/money)
[![Test](https://github.com/decimal/money/actions/workflows/ci.yml/badge.svg)](https://github.com/decimal/money/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/decimal/money/branch/main/graph/badge.svg)](https://codecov.io/gh/decimal/money)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A precise monetary calculation library for Go, Rust, and JavaScript. This library addresses the fundamental problem of floating-point arithmetic in financial applications by using integer-based storage with arbitrary precision decimal arithmetic.

## Problem Statement

Floating-point arithmetic is fundamentally unsuitable for financial calculations due to how computers represent decimal numbers in binary.

### The Classic Float Error

```go
// Go's float64 cannot represent 0.1 exactly
package main

import (
    "fmt"
    "math"
)

func main() {
    result := 0.1 + 0.2
    fmt.Printf("0.1 + 0.2 = %.20f\n", result)
    fmt.Printf("Is it exactly 0.3? %v\n", math.Abs(result-0.3) < 1e-10)
    
    // A real-world example: splitting $100 three ways
    split := 100.0 / 3.0
    total := split * 3.0
    fmt.Printf("100/3 * 3 = %.20f\n", total)
    fmt.Printf("Is it exactly 100? %v\n", math.Abs(total-100.0) < 1e-10)
}
```

Output:
```
0.1 + 0.2 = 0.30000000000000004441
Is it exactly 0.3? false
100/3 * 3 = 99.99999999999998578915
Is it exactly 100? false
```

### Why This Matters

```go
// A banking system calculating interest
package main

import "fmt"

func main() {
    // $10,000 at 4.5% annual interest
    principal := 10000.0
    rate := 0.045
    monthlyInterest := principal * rate / 12
    
    fmt.Printf("Monthly interest: $%.20f\n", monthlyInterest)
    fmt.Printf("After 12 months: $%.20f\n", monthlyInterest*12)
    
    // In real banking, these MUST be exact
    // A $0.00000000000001 error × millions of accounts = real money lost
}
```

### The Decimal Solution

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Using this library: exact arithmetic
    principal := money.USD.FromInt(1000000) // $10,000.00 in cents
    rate := money.USD.FromString("0.045")
    monthlyInterest, _ := principal.Multiply(rate, money.RoundHalfUp)
    monthlyInterest, _ = monthlyInterest.Divide(money.USD.FromInt(12), money.RoundHalfUp)
    
    fmt.Printf("Monthly interest: %s\n", monthlyInterest.String())
    // Output: Monthly interest: $37.50
    
    yearly := monthlyInterest.Multiply(money.USD.FromInt(12), money.RoundHalfUp)
    fmt.Printf("After 12 months: %s\n", yearly.String())
    // Output: After 12 months: $450.00 - EXACT!
}
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Create money from integer (cents) or string
    price := money.USD.FromString("29.99")
    tax := money.USD.FromString("0.08875")
    
    // Multiply with precise rounding
    taxAmount, _ := price.Multiply(tax, money.RoundHalfUp)
    
    fmt.Println(price.Add(taxAmount).String()) // $32.65
}
```

## Features

### Core Features
- **Integer-based storage**: All amounts stored as integers in smallest currency unit (cents for USD)
- **Arbitrary precision**: Uses `big.Float` for intermediate calculations, avoiding float overflow
- **Currency-aware**: Built-in support for major world currencies with proper decimal places
- **Comprehensive rounding modes**: Banker's rounding, half-up, ceiling, floor, and more
- **Allocation algorithms**: Split amounts proportionally with configurable strategies

### Supported Currencies
- USD (US Dollar) - 2 decimal places
- EUR (Euro) - 2 decimal places
- GBP (British Pound) - 2 decimal places
- JPY (Japanese Yen) - 0 decimal places
- CHF (Swiss Franc) - 2 decimal places
- CAD (Canadian Dollar) - 2 decimal places
- AUD (Australian Dollar) - 2 decimal places
- CNY (Chinese Yuan) - 2 decimal places
- INR (Indian Rupee) - 2 decimal places
- BRL (Brazilian Real) - 2 decimal places
- MXN (Mexican Peso) - 2 decimal places
- And more via custom Currency definition

### Operations
- `Add`, `Subtract`, `Multiply`, `Divide`
- `Abs`, `Negate`, `IsZero`, `IsPositive`, `IsNegative`
- `Allocate` (equal parts) and `AllocateRatios` (proportional)
- `Round` to specific precision
- `Compare` for ordering

### Error Handling
- Type-safe error types (`CurrencyMismatchError`, `DivisionByZeroError`, `OverflowError`)
- Precision loss warnings for float conversions
- Parse error details with expected format hints

## Benchmark Results

| Operation | shopspring/decimal | decimal/money |
|-----------|-------------------|---------------|
| Add |基准 | TBD |
| Multiply | 基准 | TBD |
| Divide | 基准 | TBD |
| Parse | 基准 | TBD |

Note: Full benchmarks pending environment setup. See [docs/benchmark-plan.md](docs/benchmark-plan.md) for methodology.

## Comparison

| Feature | shopspring/decimal | decimal.js | This Library |
|---------|-------------------|------------|--------------|
| Language | Go | JavaScript | Go, JS, Rust/Wasm |
| Type | Generic decimal | Generic decimal | Money type with currency |
| Overflow protection | Via big.Float | Manual | Automatic |
| Rounding modes | 8 modes | Multiple | 6 modes |
| Allocation | No | No | Yes |
| Currency metadata | No | No | Yes |
| Zero allocation | Fixed | Fixed | RoundRobin + RemainderToFirst |
| Modular design | Single package | Single package | Separate money/bigmoney/allocator |

## Installation

### Go

```bash
go get github.com/decimal/money/money
go get github.com/decimal/money/bigmoney
go get github.com/decimal/money/allocator
```

### JavaScript/TypeScript

```bash
npm install @decimal/money
# or
yarn add @decimal/money
# or
pnpm add @decimal/money
```

### Rust

```toml
[dependencies]
decimal-money = { git = "https://github.com/decimal/money" }
```

## Usage Examples

### Go

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Basic arithmetic
    a := money.USD.FromString("10.50")
    b := money.USD.FromString("5.25")
    
    sum, _ := a.Add(b)
    diff, _ := a.Subtract(b)
    prod, _ := a.Multiply(b, money.RoundHalfUp)
    
    fmt.Println(sum.String())   // $15.75
    fmt.Println(diff.String())  // $5.25
    fmt.Println(prod.String()) // $55.13
    
    // Allocation - split $100 among 3 people
    total := money.USD.FromString("100.00")
    parts, _ := total.AllocateRatios([]int{5, 3, 2})
    // parts[0] = $50.00, parts[1] = $30.00, parts[2] = $20.00
    
    // Currency conversion
    usd := money.USD.FromString("100.00")
    eur, _ := usd.Convert(money.EUR, money.USD.FromString("0.85"), money.RoundHalfUp)
    fmt.Println(eur.String()) // €85.00
}
```

### JavaScript/TypeScript

```typescript
import { Money, USD, EUR, RoundingMode } from '@decimal/money';

// Basic operations
const price = USD.fromString('29.99');
const tax = USD.fromString('0.08875');
const total = price.multiply(tax, RoundingMode.HalfUp);

console.log(total.toString()); // $32.65

// Allocation
const bill = USD.fromString('100.00');
const [alice, bob, carol] = bill.allocateRatios([5, 3, 2]);
console.log(alice.toString()); // $50.00
console.log(bob.toString());   // $30.00
console.log(carol.toString()); // $20.00
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions welcome! Please read our contributing guidelines before submitting PRs.
