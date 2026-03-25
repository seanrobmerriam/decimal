# ADR 0003: Currency as Type Member

## Status
Accepted

## Date
2024-01-15

## Context
In the Decimal Money Library, the `Money` type represents a monetary amount in a specific currency. The design decision we're documenting here is whether currency should be:

1. **A member of the Money type** (current approach)
2. **A generic type parameter** (e.g., `Money<C>` where C is a currency type)
3. **A separate argument** (e.g., functions take `(amount int64, currencyCode string)`)
4. **A phantom type** (currency tracked only at compile time)

## Decision
We store currency as a member field of the `Money` type:

```go
type Money struct {
    amount   int64  // Amount in smallest unit (cents)
    currency Currency
}
```

### Rationale

1. **Type Safety Within Operations**

When you call `money1.Add(money2)`, the operation can verify that both amounts are in the same currency at compile time and runtime:

```go
usd := money.New(100, "USD")
eur := money.New(100, "EUR")

// This would return an error
result, err := usd.Add(eur) // ErrCurrencyMismatch
```

With generic currency types, you could accidentally try to add `Money<USD>` to `Money<EUR>` at compile time, but the error handling would be more complex.

2. **Runtime Currency Verification**

Financial applications frequently need to validate currency at runtime from external sources (databases, user input, API responses). Having currency as a runtime member allows:

```go
// Parse from database or API
money, err := money.Parse("$100.00", "USD")

// Validate currency is supported
if !money.Currency().IsValid() {
    return ErrUnsupportedCurrency
}
```

3. **Explicit is Better Than Implicit**

Storing currency in the type makes the semantics clear:

```go
// What you see is what you get
m := money.New(100, "USD")
fmt.Println(m.Amount()) // 100 (in cents)
fmt.Println(m.Currency()) // USD
```

4. **Simpler API Surface**

A single `Money` type with a currency member is simpler than:
- Multiple generated types (MoneyUSD, MoneyEUR, etc.)
- Complex generic constraints
- Separate currency argument passing on every function

5. **JSON/Serialization Compatibility**

Modern APIs often serialize money as:
```json
{ "amount": 100, "currency": "USD" }
```

Having currency as a member makes serialization straightforward.

## Consequences

### Positive
- Clear, simple mental model
- Runtime type safety via the `Add`, `Subtract`, `Compare` operations
- Natural JSON/database serialization
- Single type to understand and document
- Consistent with standard financial libraries (Java BigDecimal with Currency, Python Decimal with implicit context)

### Negative
- Cannot have compile-time guarantee that operations use matching currencies
- Currency is stored per-instance, using more memory in large collections
- Requires runtime checks that could be caught at compile time

### Mitigations
- Implement `Add`, `Sub`, `Mul`, `Div` methods that return errors on currency mismatch
- Consider a `MoneyPair` or `MoneyBag` type for batch operations that explicitly track multiple currencies
- Document that high-performance scenarios should validate currency before bulk operations

## Alternatives Considered

### Generic Type Parameter: `Money[C Currency]`

```go
type Money[C Currency] struct {
    amount int64
}

func Add[C Currency](a, b Money[C]) Money[C] { ... }
```

**Pros**: Compile-time currency matching
**Cons**: 
- Complex type signatures
- Difficult JSON serialization
- Harder to use with dynamic currency values from databases
- Increased cognitive load

### Separate Currency Argument

```go
func Add(amount1 int64, curr1, amount2 int64, curr2 string) (int64, string, error)
```

**Pros**: No type needed
**Cons**:
- Easy to pass wrong currency
- No type safety whatsoever
- Verbose function signatures
- Not idiomatic Go

### Phantom Type

```go
type USD struct{}
type EUR struct{}

type Money struct {
    amount int64
    _      USD // phantom - no actual storage
}
```

**Pros**: Compile-time safety for known currencies
**Cons**:
- Cannot handle dynamic currencies from external sources
- Complex type hierarchy
- Serialization difficulties

## Implementation

The current implementation stores currency as:

```go
type Money struct {
    amount   int64
    currency Currency
}

type Currency struct {
    code     string
    exponent int32 // decimal places (2 for most currencies)
    name     string
}
```

## References

- [Java BigDecimal Currency](https://docs.oracle.com/javase/8/docs/api/java/math/BigDecimal.html)
- [Martin Fowler's Money Pattern](https://martinfowler.com/eaaCatalog/money.html)
- [Python Decimal with Currency Context](https://docs.python.org/3/library/decimal.html)
