# ADR-0001: Type System Design for Decimal Money Library

## Status

**Accepted** - 2024-01-15

## Context

We are building a production-grade decimal money arithmetic library for financial systems, targeting both Go (backend) and WebAssembly/JavaScript (browser) environments. The core design challenge is how to represent monetary values in a way that:

1. Eliminates floating-point errors in calculations
2. Prevents mixing incompatible currencies
3. Provides excellent performance (targeting 3-5x faster than shopspring/decimal)
4. Maintains type safety at compile time where possible
5. Handles overflow gracefully for extreme but valid financial values

### Design Forces

- **Financial accuracy**: Money calculations must be exact; no rounding errors except those explicitly requested
- **Type safety**: Operations between incompatible currencies should be compile-time or clear runtime errors
- **Performance**: Hot path operations should have zero allocations and inline well
- **Portability**: Same core logic in Go and Rust/Wasm
- **Familiarity**: API should be intuitive to developers used to `shopspring/decimal` or `decimal.js`

---

## Decision

### Primary Decision: Currency as Type Member

**Chosen Approach**: The `Money` type embeds `Currency` as a member, not as a type parameter or separate argument.

```go
type Money struct {
    amount   int64
    currency Currency
}
```

### Supporting Decisions

1. **Amount Storage**: `int64` with overflow detection (not `big.Int`) for the primary type
2. **Scale from Currency**: Decimal places determined by `Currency.exponent`, not stored per-value
3. **Separate `BigMoney` type**: For values exceeding `int64` range
4. **Separate `Decimal` type**: For currency-agnostic decimal arithmetic

---

## Detailed Analysis

### Option A: Currency as Type Member (CHOSEN)

```go
type Money struct {
    amount   int64
    currency Currency
}

type Currency struct {
    code     string
    exponent int
}
```

**Pros**:
- Self-documenting: every Money value knows its currency
- Type safety: different currencies are different types at runtime
- Clear error messages with currency context
- Simple mental model
- API chaining works naturally

**Cons**:
- Currency stored per-instance (minor memory overhead)
- Compile-time generics not leveraged (Go lacks powerful generics)
- Changing currency requires creating new Money

**Code Example**:
```go
price := money.USD.FromString("19.99")
tax, _ := price.Mul(money.Decimal.NewFloat(0.08))  // OK: same currency
_, err := price.Add(money.EUR.FromInt(1000))        // Error: currency mismatch
```

---

### Option B: Currency as Generic Type Parameter

```go
type Money[C Currency] struct {
    amount int64
}

var USDMoney = Money[USD]{}
```

**Pros**:
- Compile-time currency checking possible
- Zero-cost abstraction at runtime

**Cons**:
- Complex API: `Money[USD]` everywhere
- Poor Go idioms: Go generics are verbose
- Hard to store heterogenous money in collections
- Over-engineered for the problem

**Code Example**:
```go
price := Money[USD]{amount: 1999}
tax := price.Mul(0.08)  // Error: cannot infer type
// Requires: price.Mul[USD](0.08)
```

**Verdict**: Rejected - too complex for Go idioms

---

### Option C: Currency Passed Separately

```go
func Add(a, b Money, currency Currency) Money
func Mul(m Money, scalar decimal.Decimal, currency Currency) Money
```

**Pros**:
- Simple Money type (just amount)
- Flexible currency handling

**Cons**:
- Runtime currency errors only
- No compile-time safety
- API feels untyped
- Error-prone: easy to forget currency parameter

**Code Example**:
```go
price := Money{amount: 1999}
Add(price, tax, USD)  // OK
Add(price, tax, EUR)  // Runtime error if wrong
```

**Verdict**: Rejected - type safety is paramount for financial code

---

### Amount Storage: int64 vs big.Int

#### Option A: int64 (CHOSEN)

```go
type Money struct {
    amount int64  // Stored in smallest unit (cents for USD)
}
```

**Pros**:
- Excellent performance (single instruction add on 64-bit systems)
- Cache-friendly (16 bytes fits in one cache line)
- Zero allocations for operations
- Memory efficient

**Cons**:
- Limited range: ±9.2 quintillion (sufficient for most financial uses)
- Overflow possible for extreme calculations

**Range Analysis**:
```
Max int64: 9,223,372,036,854,775,807
As USD cents: $92,233,720,368,547,759.57 (~$92 quadrillion)
As JPY (0 decimals): ¥9,223,372,036,854,775,807

For context:
- US GDP: ~$25 trillion
- World GDP: ~$100 trillion
- US national debt: ~$34 trillion
```

**Verdict**: int64 is sufficient for 99.99% of financial use cases

---

#### Option B: big.Int

```go
type Money struct {
    amount *big.Int  // Heap-allocated arbitrary precision
}
```

**Pros**:
- No overflow concerns
- Truly unlimited range

**Cons**:
- Heap allocation on every operation
- 10-50x slower than int64
- GC pressure
- Complex memory management

**Verdict**: Rejected for primary type; `BigMoney` available for overflow cases

---

## Consequences

### Positive

1. **Type Safety**: Currency mismatch errors at runtime with clear context
2. **Performance**: Zero allocations in hot paths, excellent CPU cache usage
3. **Simplicity**: Mental model is straightforward
4. **Interoperability**: Clear separation between Money (with currency) and Decimal (without)
5. **Overflow Handling**: Clear path forward (BigMoney) when int64 is insufficient

### Negative

1. **Memory Overhead**: ~16 bytes per Money vs 8 bytes for pure amount
2. **Generic Limitations**: Cannot achieve compile-time currency checking in Go
3. **Currency Change**: Must create new Money to change currency

### Mitigation

1. **Memory**: 16 bytes is still cache-line friendly; memory overhead is negligible
2. **Compile-time Safety**: Achieved through code review and error handling
3. **API**: Provide `money.Convert()` for currency changes

---

## Implementation Details

### Money Type Definition

```go
// Money represents a monetary amount with a specific currency.
// Amount is stored as an integer in the currency's smallest unit (cents for USD).
type Money struct {
    amount   int64     // Amount in smallest currency unit
    currency Currency  // ISO 4217 currency code and properties
}

// Currency represents an ISO 4217 currency.
type Currency struct {
    code     string  // ISO 4217 code (e.g., "USD")
    exponent int     // Number of decimal places (e.g., 2 for USD)
    name     string  // Full name (e.g., "United States Dollar")
}
```

### Decimal Type (Currency-Agnostic)

```go
// Decimal represents an arbitrary-precision decimal number without currency.
// Used for intermediate calculations, percentages, and non-monetary values.
type Decimal struct {
    value *big.Int  // Unscaled integer value
    scale int32     // Number of decimal places
}
```

### BigMoney (Overflow Case)

```go
// BigMoney provides arbitrary precision for extreme values.
type BigMoney struct {
    amount   *big.Int  // Arbitrary precision
    currency Currency
}
```

### Type Hierarchy

```
Decimal (currency-agnostic)
├── Operations: add, sub, mul, div, round, etc.
└── Used for: scalars, rates, intermediate values

Money (currency-specific, int64)
├── Operations: + - * / (with Currency type safety)
├── Error: CurrencyMismatch when currencies differ
└── Used for: primary monetary values

BigMoney (currency-specific, big.Int)
├── Operations: same as Money
├── Overflow: automatic promotion from Money operations
└── Used for: values exceeding int64 range
```

---

## Code Examples

### Creating Money

```go
// From integer cents
price := money.USD.FromInt(1999)  // $19.99

// From string
price, err := money.USD.FromString("19.99")  // $19.99

// Zero
zero := money.USD.Zero()

// From float (WARNING: may lose precision)
price, err := money.USD.FromFloat(19.99)  // approximate
```

### Arithmetic with Type Safety

```go
usd1 := money.USD.FromInt(1000)
usd2 := money.USD.FromInt(500)

sum, err := usd1.Add(usd2)  // OK: $15.00

_, err := usd1.Add(money.EUR.FromInt(500))  // Error: CurrencyMismatchError
```

### Currency Conversion (Explicit)

```go
usd := money.USD.FromInt(10000)  // $100.00
rate := money.Decimal.NewFloat(0.92)

// Explicit conversion with rate
eur, err := usd.Convert(money.EUR, rate, money.RoundHalfEven)
// eur: €92.00
```

### Using Decimal for Rates

```go
price := money.USD.FromInt(10000)  // $100.00
taxRate := money.Decimal.NewFloat(0.0825)

// Multiply returns Money (preserves currency)
tax, err := price.Mul(taxRate)
// tax: $8.25

// Round to 2 decimal places
tax = tax.Round(money.RoundHalfEven, 2)
```

### Handling Overflow

```go
// int64 can handle most values
price := money.USD.FromInt(math.MaxInt64 - 1)

// Multiplication that exceeds int64
bigPrice := money.USD.FromInt(math.MaxInt64)
huge, err := bigPrice.Mul(2)  // Error: OverflowError

// Use BigMoney for extreme values
bigMoney := money.ParseBigMoney("USD", "9223372036854775807000")
hugeMoney, _ := bigMoney.Mul(2)  // OK
```

---

## Alternatives Considered

### 1. Currency Code in String

```go
type Money struct {
    amount   int64
    currency string  // "USD" instead of Currency struct
}
```

**Rejected**: Loses exponent (decimal places) information; requires lookup for every operation.

### 2. Tagged Type Pattern

```go
type USDMoney struct {
    cents int64
}
type EURMoney struct {
    cents int64
}
```

**Rejected**: N distinct types for N currencies; operations between currencies require explicit conversion; complex API.

### 3. Decimal Places Per-Instance

```go
type Money struct {
    amount int64
    scale  int  // Stored per-instance
}
```

**Rejected**: Same currency can have different scales; adds complexity without benefit; currency already defines scale.

---

## Related Decisions

- **ADR-0002**: Rounding Modes (RoundHalfEven as default)
- **ADR-0003**: Error Handling Strategy (typed errors with context)
- **ADR-0004**: Wasm Memory Model (value types where possible)

---

## References

- [ISO 4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html)
- [IEEE 754 Rounding Recommendations](https://ieeexplore.ieee.org/document/8767899)
- [POSIX Money Calculations](https://pubs.opengroup.org/onlinepubs/9699919799/functions/printf.html)
- [shopspring/decimal](https://github.com/shopspring/decimal) - Influencing library
- [decimal.js](https://github.com/MikeMcl/decimal.js) - JavaScript reference

---

## Review History

| Date | Version | Author | Changes |
|------|---------|--------|---------|
| 2024-01-15 | 1.0 | Architecture Team | Initial decision |
| 2024-01-16 | 1.1 | Code Team | Added BigMoney for overflow |
| 2024-02-01 | 1.2 | Security Review | Confirmed type safety adequate |

---

## Appendix: Type Size Analysis

```
┌─────────────────────┬────────────┬─────────────────────────────────┐
│ Type                │ Size       │ Notes                           │
├─────────────────────┼────────────┼─────────────────────────────────┤
│ Money (chosen)      │ 16 bytes   │ 8 amount + 3 code + 1 exp + 4 pad│
│ Money (string curr) │ 24 bytes   │ 8 amount + 16 string            │
│ Money (tagged)      │ 8 bytes    │ Per-currency type overhead      │
│ Decimal             │ 24 bytes   │ 16 big.Int pointer + 4 scale    │
│ BigMoney            │ 40 bytes   │ 32 big.Int + 8 currency         │
└─────────────────────┴────────────┴─────────────────────────────────┘

Memory per 1M Money values:
- Chosen approach: 16 MB
- BigInt approach: 24+ MB + GC pressure
- Difference: ~8MB per million values (acceptable)