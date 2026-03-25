# WebAssembly (Rust/Wasm) Binding Strategy

## Overview

This document describes the design for exposing the decimal money library to JavaScript via WebAssembly compiled from Rust. The goal is to provide a high-performance, type-safe decimal arithmetic library for browser-based financial applications.

---

## Why Rust for Wasm?

### Rationale

1. **Performance**: Rust produces highly optimized Wasm bytecode with minimal runtime overhead
2. **Safety**: Memory safety without garbage collection pauses
3. **Bundle Size**: Smaller Wasm binaries compared to other compiled languages
4. **Ecosystem**: Excellent Wasm tooling (wasm-pack, wasm-bindgen)
5. **Interoperability**: Seamless FFI with JavaScript via wasm-bindgen

### Alternatives Considered

| Language | Pros | Cons |
|----------|------|------|
| **Rust (CHOSEN)** | Best optimization, no GC, small binaries | Steeper learning curve |
| Go | Familiar to Go developers | Larger binaries, GC pauses |
| C/C++ | Mature toolchain | Manual memory management risk |
| AssemblyScript | TypeScript compatible | Limited optimization, younger project |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      JavaScript Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Money class │  │ Decimal.js  │  │ Your Application     │ │
│  │ (JS API)    │  │ compat      │  │                     │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────────────┘ │
└─────────┼────────────────┼─────────────────────────────────┘
          │                │
          ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                   wasm-bindgen Layer                        │
│  ┌────────────────────────────────────────────────────────┐│
│  │  #[wasm_bindgen] public API surface                    ││
│  │  - Type conversions                                     ││
│  │  - Error handling (JS Error objects)                    ││
│  │  - Memory management                                    ││
│  └────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                      Rust Core                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Money type  │  │ Decimal type│  │ Internal algorithms │ │
│  │ (same as Go)│  │ (same as Go)│  │ (same as Go)        │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Memory Layout

### JavaScript-Wasm Memory Model

```rust
// Money value represented as two i32 values in linear memory
// This allows passing by value (not by reference) for performance

#[repr(C)]
pub struct WasmMoney {
    amount_low: i32,    // Lower 32 bits of amount
    amount_high: i32,   // Upper 32 bits of amount (for overflow cases)
    currency: i32,      // Currency code index (small integer)
}

// For most cases, we can use simpler representation
#[repr(C)]
pub struct WasmMoneySimple {
    amount: i64,       // Single int64 for normal range
    currency: i32,      // Currency code index
}
```

### Zero-Copy Considerations

**Problem**:wasm-bindgen copies data when passing between JS and Wasm.

**Solution**: For bulk operations, use bulk memory access:

```rust
// Instead of returning large arrays, write to shared memory
#[wasm_bindgen]
pub fn split_allocate(
    money_ptr: *const WasmMoney,
    n: i32,
    results_ptr: *mut WasmMoney,
    capacity: i32
) -> i32 {
    // Write results directly to Wasm memory
    // JS reads from memory without additional copy
}
```

---

## Public API Surface

### Module Structure

```rust
// src/lib.rs
pub mod decimal;
pub mod money;
pub mod errors;

pub use decimal::Decimal;
pub use money::{Money, Currency, RoundingMode};
pub use errors::{Error, ErrorKind};
```

### Money Type

```rust
use wasm_bindgen::prelude::*;
use crate::errors::{Error, ErrorKind};

// #[wasm_bindgen] exposes this to JavaScript
#[wasm_bindgen]
#[derive(Clone, Debug, PartialEq)]
pub struct Money {
    inner: decimal::Money,
}

// Money construction
#[wasm_bindgen]
impl Money {
    /// Create Money from integer amount in smallest currency unit.
    /// Example: Money.fromInt("USD", 1999) = $19.99
    #[wasm_bindgen(constructor)]
    pub fn from_int(currency: &str, amount: f64) -> Result<Money, Error> {
        let currency = parse_currency(currency)?;
        let amount = amount as i64;  // f64 can represent all i64 values
        Ok(Money { inner: currency.from_int(amount) })
    }
    
    /// Create Money from string representation.
    /// Example: Money.fromString("USD", "19.99") = $19.99
    #[wasm_bindgen]
    pub fn from_string(currency: &str, value: &str) -> Result<Money, Error> {
        let currency = parse_currency(currency)?;
        currency.from_string(value)
            .map(|m| Money { inner: m })
            .map_err(|e| Error::from(e))
    }
    
    /// Create Money from float (WARNING: may lose precision)
    #[wasm_bindgen]
    pub fn from_float(currency: &str, value: f64) -> Result<Money, Error> {
        let currency = parse_currency(currency)?;
        Ok(Money { inner: currency.from_float(value) })
    }
    
    // Binary operations
    #[wasm_bindgen]
    pub fn add(&self, other: &Money) -> Result<Money, Error> {
        self.inner.add(&other.inner)
            .map(|m| Money { inner: m })
            .map_err(|e| Error::from(e))
    }
    
    #[wasm_bindgen]
    pub fn sub(&self, other: &Money) -> Result<Money, Error> {
        self.inner.sub(&other.inner)
            .map(|m| Money { inner: m })
            .map_err(|e| Error::from(e))
    }
    
    #[wasm_bindgen]
    pub fn mul(&self, scalar: &Decimal) -> Result<Money, Error> {
        self.inner.mul(scalar.as_ref())
            .map(|m| Money { inner: m })
            .map_err(|e| Error::from(e))
    }
    
    #[wasm_bindgen]
    pub fn div(&self, scalar: &Decimal, rounding: RoundingMode) -> Result<Money, Error> {
        self.inner.div(scalar.as_ref(), rounding.into())
            .map(|m| Money { inner: m })
            .map_err(|e| Error::from(e))
    }
    
    // Comparison
    #[wasm_bindgen]
    pub fn eq(&self, other: &Money) -> bool {
        self.inner == other.inner
    }
    
    #[wasm_bindgen]
    pub fn gt(&self, other: &Money) -> Result<bool, Error> {
        Ok(self.inner.cmp(&other.inner).is_gt())
    }
    
    #[wasm_bindgen]
    pub fn lt(&self, other: &Money) -> Result<bool, Error> {
        Ok(self.inner.cmp(&other.inner).is_lt())
    }
    
    // Allocation
    #[wasm_bindgen]
    pub fn split(&self, n: u32) -> Result<Vec<Money>, Error> {
        self.inner.split(n as usize)
            .map(|v| v.into_iter().map(|m| Money { inner: m }).collect())
            .map_err(|e| Error::from(e))
    }
    
    // Representation
    #[wasm_bindgen]
    pub fn to_string(&self) -> String {
        self.inner.to_string()
    }
    
    #[wasm_bindgen]
    pub fn to_float(&self) -> f64 {
        self.inner.to_float()
    }
    
    #[wasm_bindgen]
    pub fn currency(&self) -> String {
        self.inner.currency().code().to_string()
    }
}
```

---

## JavaScript API Design

### Option 1: Direct Class API (Chosen)

Mirrors Rust API as closely as possible:

```javascript
import init, { Money, Decimal, RoundingMode } from './decimal_wasm.js';

await init();  // Initialize Wasm module

// Construction
const price = new Money("USD", "19.99");
const tax = Money.fromInt("USD", 150);  // $1.50

// Operations
const total = price.add(tax);
console.log(total.toString());  // "$21.49"

// With BigInt for large amounts
const largeAmount = Money.fromString("USD", "1234567890123456.78");
```

### Option 2: Builder Pattern

```javascript
const result = Money.of("USD", 1000)
    .add(Money.of("USD", 500))
    .mul(Decimal.new("1.05"))
    .round(RoundingMode.HALF_EVEN)
    .toString();
```

### Option 3: decimal.js Compatibility Layer

For migration from existing decimal.js:

```javascript
// Wrapper that provides decimal.js-like API
class DecimalMoney {
    constructor(value, currency = 'USD') {
        this.money = Money.fromString(currency, value.toString());
    }
    
    plus(other) {
        return new DecimalMoney(this.money.add(other.money).toString());
    }
    
    minus(other) {
        return new DecimalMoney(this.money.sub(other.money).toString());
    }
    
    times(scalar) {
        return new DecimalMoney(this.money.mul(Decimal.fromString(scalar.toString())).toString());
    }
    
    dividedBy(scalar) {
        return new DecimalMoney(
            this.money.div(Decimal.fromString(scalar.toString()), RoundingMode.HALF_EVEN).toString()
        );
    }
    
    toFixed(decimals) {
        return this.money.round(RoundingMode.HALF_EVEN, decimals).toString();
    }
}
```

---

## Error Handling

### Strategy: JavaScript Error Objects

Convert Rust errors to JavaScript errors for idiomatic JS handling:

```rust
#[derive(Debug)]
pub enum ErrorKind {
    CurrencyMismatch,
    DivisionByZero,
    Overflow,
    InvalidCurrency,
    ParseError,
    RoundingError,
}

#[wasm_bindgen]
pub struct Error {
    kind: ErrorKind,
    message: String,
}

#[wasm_bindgen]
impl Error {
    #[wasm_bindgen(getter)]
    pub fn kind(&self) -> String {
        format!("{:?}", self.kind)
    }
    
    #[wasm_bindgen(getter)]
    pub fn message(&self) -> String {
        self.message.clone()
    }
}

impl From<decimal::Error> for Error {
    fn from(err: decimal::Error) -> Error {
        match err.kind() {
            decimal::ErrorKind::CurrencyMismatch => Error {
                kind: ErrorKind::CurrencyMismatch,
                message: err.to_string(),
            },
            // ... other conversions
        }
    }
}
```

### JavaScript Usage

```javascript
try {
    const result = usd.add(eur);
} catch (e) {
    if (e.kind === 'CurrencyMismatch') {
        console.error('Cannot add USD and EUR');
    } else if (e.kind === 'DivisionByZero') {
        console.error('Division by zero');
    }
}
```

---

## Performance Optimizations

### 1. Inline Operations

Use `#[inline(always)]` for small functions:

```rust
#[inline(always)]
pub fn amount(&self) -> i64 {
    self.inner.amount()
}

#[inline(always)]
pub fn is_zero(&self) -> bool {
    self.inner.is_zero()
}
```

### 2. Lazy Static Initialization

```rust
use once_cell::sync::Lazy;
use std::collections::HashMap;

static CURRENCY_MAP: Lazy<HashMap<String, Currency>> = Lazy::new(|| {
    let mut m = HashMap::new();
    m.insert("USD".to_string(), Currency::USD);
    m.insert("EUR".to_string(), Currency::EUR);
    // ...
    m
});
```

### 3. Bulk Operations

```rust
#[wasm_bindgen]
pub fn add_all(monies: &[Money]) -> Result<Money, Error> {
    if monies.is_empty() {
        return Err(Error::new("Empty list"));
    }
    
    let mut sum = monies[0].inner.clone();
    for money in &monies[1..] {
        sum = sum.add(&money.inner)?;
    }
    Ok(Money { inner: sum })
}
```

### 4. Memory Pre-allocation

```rust
// Pre-allocate result vectors to avoid reallocation
#[wasm_bindgen]
pub fn split(&self, n: usize) -> Result<Vec<Money>, Error> {
    let mut results = Vec::with_capacity(n);
    // ... populate
    results
}
```

---

## Bundle Size Optimization

### 1. wasm-opt

```bash
# Apply Wasm optimizations
wasm-opt -O3 --zero-filled-memory --merge-blocks --remove-unused-br \
    target/wasm32-unknown-unknown/release/decimal_wasm.wasm \
    -o decimal_wasm_opt.wasm
```

### 2. Binary Size Comparison

| Library | Minified + Gzipped |
|---------|-------------------|
| decimal.js | ~21 KB |
| big.js | ~10 KB |
| This library (Rust/Wasm) | ~15 KB (target) |
| shopspring/decimal (Go) | N/A (not for browser) |

### 3. Tree Shaking

```javascript
// Only import what you need
import { Money } from './decimal_wasm.js';
// NOT: import * as decimal from './decimal_wasm.js';
```

### 4. Feature Flags

```rust
[features]
default = ["std", "serde"]
std = []
serde = ["dep:serde"]
# Disable for smaller bundle
big = []  # Use big-int for overflow cases
```

---

## Type Conversions

### Rust → JavaScript

```rust
#[wasm_bindgen]
impl Money {
    // Rust String → JS String
    #[wasm_bindgen(method)]
    pub fn to_string(&self) -> String { ... }
    
    // i64 → f64 (lossy for very large numbers)
    #[wasm_bindgen(method)]
    pub fn to_float(&self) -> f64 { ... }
    
    // Return as BigInt for precision preservation
    #[wasm_bindgen(method, js_name = "toBigInt")]
    pub fn to_big_int(&self) -> js_sys::BigInt { ... }
}
```

### JavaScript → Rust

```rust
#[wasm_bindgen]
impl Money {
    // JS String → Rust String (handled by wasm-bindgen)
    pub fn from_string(currency: &str, value: &str) -> Result<Money, Error> { ... }
    
    // JS Number → i64 (may lose precision for very large numbers)
    pub fn from_float(currency: &str, value: f64) -> Result<Money, Error> { ... }
    
    // BigInt support
    pub fn from_big_int(currency: &str, value: &js_sys::BigInt) -> Result<Money, Error> {
        let n = i64::try_from(value)
            .map_err(|_| Error::new("BigInt out of range"))?;
        Ok(Money { inner: currency.from_int(n)? })
    }
}
```

---

## Testing Strategy

### Unit Tests (Rust)

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_money_add() {
        let usd = Money::from_int("USD", 1000).unwrap();
        let other = Money::from_int("USD", 500).unwrap();
        let result = usd.add(&other).unwrap();
        assert_eq!(result.to_string(), "$15.00");
    }
    
    #[test]
    fn test_currency_mismatch() {
        let usd = Money::from_int("USD", 1000).unwrap();
        let eur = Money::from_int("EUR", 1000).unwrap();
        assert!(usd.add(&eur).is_err());
    }
}
```

### Property-Based Tests

```rust
#[cfg(test)]
mod property_tests {
    use proptest::prelude::*;
    
    proptest! {
        #[test]
        fn test_add_sub_inverse(a: i64, b: i64) {
            let usd = Currency::USD;
            let m1 = usd.from_int(a);
            let m2 = usd.from_int(b);
            
            let sum = m1.add(&m2)?;
            let back = sum.sub(&m2)?;
            
            prop_assert_eq!(m1.amount(), back.amount());
        }
    }
}
```

### Integration Tests (JavaScript)

```javascript
import { Money, RoundingMode } from './decimal_wasm_bg.wasm';

describe('Money', () => {
    it('should add correctly', async () => {
        const a = new Money("USD", "10.00");
        const b = new Money("USD", "5.00");
        const sum = a.add(b);
        expect(sum.toString()).toBe("$15.00");
    });
    
    it('should throw on currency mismatch', async () => {
        const a = new Money("USD", "10.00");
        const b = new Money("EUR", "5.00");
        expect(() => a.add(b)).toThrow();
    });
});
```

---

## Build Pipeline

```yaml
# .github/workflows/wasm.yml
name: Wasm Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Rust
        uses: dtolnay/rust-toolchain@stable
        with:
          targets: wasm32-unknown-unknown
      
      - name: Install wasm-pack
        run: cargo install wasm-pack
      
      - name: Build
        run: wasm-pack build --target web --release
      
      - name: Optimize
        run: |
          wasm-opt -O3 target/pkg/decimal_wasm_bg.wasm \
            -o target/pkg/decimal_wasm_opt.wasm
      
      - name: Size Report
        run: |
          echo "Wasm size: $(wc -c < target/pkg/decimal_wasm_opt.wasm) bytes"
          echo "Gzipped: $(gzip -c target/pkg/decimal_wasm_opt.wasm | wc -c) bytes"
```

---

## CDN and Distribution

### Option 1: npm Package

```json
{
  "name": "decimal-wasm",
  "version": "1.0.0",
  "main": "decimal_wasm.js",
  "module": "decimal_wasm.esm.js",
  "types": "decimal_wasm.d.ts",
  "exports": {
    ".": {
      "import": "./decimal_wasm.js",
      "require": "./decimal_wasm.cjs.js"
    }
  },
  "files": [
    "decimal_wasm_bg.wasm",
    "decimal_wasm.js"
  ]
}
```

### Option 2: Direct CDN

```html
<script type="module">
    import init, { Money } from 'https://cdn.example.com/decimal-wasm@1.0.0/decimal_wasm.js';
    await init();
    
    const price = new Money("USD", "19.99");
</script>
```

---

## Migration from decimal.js/big.js

### Compatibility Functions

```rust
#[wasm_bindgen]
impl Money {
    /// Parse decimal.js style string
    #[wasm_bindgen(js_name = "fromDecimalJs")]
    pub fn from_decimal_js(value: &str, currency: &str) -> Result<Money, Error> {
        // decimal.js uses different notation (e.g., "0.1" not ".1")
        let normalized = normalize_decimal_js_string(value);
        Self::from_string(currency, &normalized)
    }
}

fn normalize_decimal_js_string(s: &str) -> String {
    // Handle decimal.js special values
    s.replace("Infinity", "")
     .replace("NaN", "")
     .replace("e", "E")
}
```

### Comparison Table

| decimal.js | This Library |
|------------|--------------|
| `new Decimal("1.5")` | `new Decimal("1.5")` |
| `decimal.plus(other)` | `decimal.add(other)` |
| `decimal.minus(other)` | `decimal.sub(other)` |
| `decimal.times(other)` | `decimal.mul(other)` |
| `decimal.dividedBy(other)` | `decimal.div(other)` |
| `decimal.toFixed(2)` | `decimal.round(RoundHalfEven, 2).toString()` |
| N/A | `new Money("USD", "1.50")` |

---

## Summary

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Source language | Rust | Performance, no GC, small binaries |
| API style | Direct class | Type safety, idiomatic |
| Error handling | JS Error objects | Idiomatic JavaScript |
| Memory model | Value types where possible | Minimize cross-boundary copies |
| Bundle format | ES modules + Wasm | Modern browser support |
| Compatibility | decimal.js wrapper | Migration path |
| Optimization | wasm-opt + feature flags | Minimize bundle size |