# Decimal Money Library

[![Go Reference](https://pkg.go.dev/badge/github.com/decimal/money.svg)](https://pkg.go.dev/github.com/decimal/money)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A precise monetary calculation library for Go, Rust, and JavaScript. Uses integer-based storage in the currency's smallest unit (e.g., cents for USD) to avoid the rounding errors inherent in floating-point arithmetic.

## The Problem

```go
0.1 + 0.2 = 0.30000000000000004441  // Not exactly 0.3
100.0 / 3.0 * 3.0 = 99.99999999999998578915  // Not exactly 100
```

Floating-point cannot represent most decimal fractions exactly. For banking, tax, and commerce, every cent must be exact.

## Quick Start

### Go

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    price := money.USD.FromString("29.99")
    rate := money.MustDecimalFromString("0.08875")

    tax, _ := price.Mul(rate)
    total, _ := price.Add(tax)

    fmt.Println(total) // USD 32.65
}
```

### Rust

```rust
use decimal_money::prelude::*;

fn main() {
    let price = Money::from_string("USD", "29.99").unwrap();
    let rate = Decimal::from_string("0.08875").unwrap();
    let total = price.mul(&rate).unwrap();
    println!("{}", total.to_string()); // USD 32.65
}
```

### JavaScript

```javascript
const { Money, Decimal } = require('decimal-money-wasm');

const price = Money.fromString("USD", "29.99");
const rate = Decimal.fromString("0.08875");
const total = price.mul(rate);
console.log(total.toString()); // USD 32.65
```

## Features

- **Integer-based storage** — amounts stored as integers in the currency's smallest unit
- **Arbitrary precision** — `math/big.Int` for intermediate calculations (Go), `bigint` (JS), `i64` (Rust)
- **Currency-aware** — 19 pre-defined ISO 4217 currencies; arithmetic validates matching currencies
- **7 rounding modes** — Up, Down, HalfUp, HalfDown, HalfEven (banker's), Ceiling, Floor
- **Allocation** — `Split(n)` for equal division, `AllocateRatios(ratios)` for proportional split
- **Multi-language** — Go (most mature), Rust (WASM target), JavaScript/TypeScript

## Supported Currencies

| Code | Currency | Decimals |
|------|----------|----------|
| USD | US Dollar | 2 |
| EUR | Euro | 2 |
| GBP | British Pound | 2 |
| JPY | Japanese Yen | 0 |
| CHF | Swiss Franc | 2 |
| CAD | Canadian Dollar | 2 |
| AUD | Australian Dollar | 2 |
| CNY | Chinese Yuan | 2 |
| INR | Indian Rupee | 2 |
| BRL | Brazilian Real | 2 |
| MXN | Mexican Peso | 2 |
| KRW | South Korean Won | 0 |
| SGD | Singapore Dollar | 2 |
| HKD | Hong Kong Dollar | 2 |
| NOK | Norwegian Krone | 2 |
| SEK | Swedish Krona | 2 |
| DKK | Danish Krone | 2 |
| NZD | New Zealand Dollar | 2 |
| ZAR | South African Rand | 2 |

## API

All three language implementations share the same API shape:

| Operation | Go | Rust | JS |
|-----------|----|------|----|
| From string | `c.FromString(s)` | `Money::from_string(c, s)` | `Money.fromString(c, s)` |
| From int (minor units) | `c.FromInt(n)` | `Money::from_int(c, n)` | `Money.fromInt(c, n)` |
| Add | `a.Add(b)` | `a.add(&b)` | `a.add(b)` |
| Subtract | `a.Sub(b)` | `a.sub(&b)` | `a.sub(b)` |
| Multiply (by Decimal) | `a.Mul(factor)` | `a.mul(&factor)` | `a.mul(factor)` |
| Divide (by Decimal) | `a.Div(divisor, rounding)` | `a.div(&divisor, rounding)` | `a.div(divisor, rounding)` |
| Multiply (by int) | `a.MulInt(n)` | — | — |
| Divide (by int) | `a.DivInt(n, rounding)` | — | — |
| Split (equal parts) | `a.Split(n)` | `a.split(n)` | `a.split(n)` |
| Allocate by ratios | `a.AllocateRatios(r)` | — | `a.allocateRatios(r)` |
| Compare | `a.Cmp(b)` | — | `a.compareTo(b)` |
| Negate | `a.Neg()` | — | — |
| Absolute | `a.Abs()` | — | — |
| Format | `a.Format()`, `a.String()` | `a.to_string()` | `a.format()`, `a.toString()` |
| Is zero/positive/negative | `a.IsZero()`, etc. | `a.is_zero()`, etc. | `a.isZero()`, etc. |

## Rounding Modes

| Mode | Direction | 1.015 → 0.01 | 1.025 → 0.01 | -1.015 → -0.01 |
|------|-----------|-------------|-------------|----------------|
| **Up** | Away from zero | 1.02 | 1.03 | -1.02 |
| **Down** | Toward zero | 1.01 | 1.02 | -1.01 |
| **HalfUp** | ≥ 0.5 rounds up | 1.02 | 1.03 | -1.02 |
| **HalfDown** | > 0.5 rounds up | 1.01 | 1.02 | -1.01 |
| **HalfEven** | Ties → nearest even | 1.02 | 1.02 | -1.02 |
| **Ceiling** | Toward +∞ | 1.02 | 1.03 | -1.01 |
| **Floor** | Toward -∞ | 1.01 | 1.02 | -1.02 |

## Error Handling

The Go implementation provides typed errors for programmatic handling:

- `CurrencyMismatchError` — different currencies in same operation
- `DivisionByZeroError` — divisor is zero
- `OverflowError` — result exceeds int64 range
- `ParseError` — invalid string format
- `PrecisionLossError` — float conversion loses precision

Rust returns `Result<T, Error>` with `ErrorKind` variants. JS throws `Error` with descriptive messages.

## Comparison

| Feature | shopspring/decimal | This Library |
|---------|-------------------|--------------|
| Language | Go | Go, JS, Rust/Wasm |
| Type | Generic decimal | Money type with currency |
| Overflow protection | Via big.Float | int64 with big.Int intermediates |
| Rounding modes | 8 modes | 7 modes |
| Allocation | No | Yes |
| Currency metadata | No | Yes |
| Modular design | Single package | money / bigmoney / allocator |

## Installation

### Go

```bash
go get github.com/decimal/money/money
go get github.com/decimal/money/bigmoney
go get github.com/decimal/money/allocator
```

### JavaScript/TypeScript

```bash
npm install decimal-money-wasm
```

### Rust

```toml
[dependencies]
decimal-money = { git = "https://github.com/decimal/money" }
```

The Rust crate compiles to WASM via `wasm-bindgen` for browser use.

## Usage Examples

### Go — Allocation

```go
total := money.USD.FromString("100.00")
parts, _ := total.Split(3)
// parts[0] = $33.34, parts[1] = $33.33, parts[2] = $33.33

ratios, _ := total.AllocateRatios([]int{5, 3, 2})
// ratios[0] = $50.00, ratios[1] = $30.00, ratios[2] = $20.00
```

### Go — Decimal arithmetic

```go
d1 := money.MustDecimalFromString("10.50")
d2 := money.MustDecimalFromString("3.50")
sum := d1.Add(d2)          // 14.00
diff := d1.Sub(d2)         // 7.00
prod := d1.Mul(d2)         // 36.7500
quot, _ := d1.Div(d2, money.RoundHalfUp)  // 3.00
```

### Rust

```rust
let a = Money::from_string("USD", "100.00").unwrap();
let b = Money::from_string("USD", "33.00").unwrap();
let sum = a.add(&b).unwrap();
assert_eq!(sum.amount(), 13300);
```

### JavaScript

```javascript
const price = Money.fromString("USD", "29.99");
const rate = Decimal.fromString("0.08");
const total = price.mul(rate);
console.log(total.toString()); // USD 2.40
```

## License

MIT License — see [LICENSE](LICENSE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
