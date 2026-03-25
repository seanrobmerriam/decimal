# Currency Handling Design

## Overview

Currency handling is one of the most critical aspects of a money library. This document describes the design decisions, implementation strategies, and safeguards for currency operations.

---

## Design Decision: Currency as Type Member

### Decision: Currency is Part of the Money Type

**Chosen Approach**: Currency is embedded in the `Money` type, making currency mismatch a compile-time or runtime type error.

```go
type Money struct {
    amount   int64
    currency Currency
}
```

### Rationale

**Compile-Time Safety Benefits**:
```go
// This won't compile - USD and EUR are different types
func calculateTotal(items []Money) Money {
    total := USD.Zero()
    for _, item := range items {
        total = total.Add(item)  // COMPILE ERROR if types differ
    }
    return total
}
```

**Runtime Safety with Clear Errors**:
```go
// If type aliasing were used, this would silently produce wrong results
usd := money.USD.FromInt(1000)  // $10.00
eur := money.EUR.FromInt(1000)  // €10.00

result, err := usd.Add(eur)
if err != nil {
    // err: CurrencyMismatchError{Expected: "USD", Got: "EUR"}
}
```

### Alternatives Considered

**Option A: Currency as Type Member (CHOSEN)**
- Self-documenting code
- Type system enforces correctness
- Clear error messages
- Slightly more verbose API

**Option B: Currency Passed Separately**
```go
// Alternative API style
func Add(a, b Money, currency Currency) (Money, error)
```
- More flexible API
- Runtime errors only
- Higher risk of bugs
- **REJECTED**: Type safety is paramount for financial code

**Option C: Currency as Generic Type Parameter**
```go
type Money[C Currency] struct { ... }
```
- Compile-time currency tracking
- Complex API, poor readability
- Poor Go idiomatic style
- **REJECTED**: Over-engineered for use case

---

## Currency Type Definition

```go
// Currency represents a currency with ISO 4217 properties.
type Currency struct {
    code     string  // ISO 4217 code: "USD", "EUR", etc.
    exponent int     // Number of decimal places: 2 for USD, 0 for JPY
    name     string  // Human-readable name
}

// String returns the currency code.
func (c Currency) String() string { return c.code }

// DecimalPlaces returns the number of decimal places.
func (c Currency) DecimalPlaces() int { return c.exponent }
```

---

## Predefined Currencies

### Major Currencies

```go
var (
    // USD - United States Dollar
    // Standard: 2 decimal places
    // Common in: United States, Ecuador, El Salvador, etc.
    USD = Currency{
        code:     "USD",
        exponent: 2,
        name:     "United States Dollar",
    }
    
    // EUR - Euro
    // Standard: 2 decimal places
    // Common in: Eurozone (19 countries)
    EUR = Currency{
        code:     "EUR",
        exponent: 2,
        name:     "Euro",
    }
    
    // GBP - British Pound Sterling
    // Standard: 2 decimal places
    // Common in: United Kingdom
    GBP = Currency{
        code:     "GBP",
        exponent: 2,
        name:     "British Pound Sterling",
    }
    
    // JPY - Japanese Yen
    // Standard: 0 decimal places (no sub-unit)
    // Common in: Japan
    JPY = Currency{
        code:     "JPY",
        exponent: 0,
        name:     "Japanese Yen",
    }
    
    // CHF - Swiss Franc
    // Standard: 2 decimal places
    // Common in: Switzerland, Liechtenstein
    CHF = Currency{
        code:     "CHF",
        exponent: 2,
        name:     "Swiss Franc",
    }
)
```

### Additional Currencies

```go
var (
    // CAD - Canadian Dollar
    CAD = Currency{code: "CAD", exponent: 2, name: "Canadian Dollar"}
    
    // AUD - Australian Dollar  
    AUD = Currency{code: "AUD", exponent: 2, name: "Australian Dollar"}
    
    // CNY - Chinese Yuan (Renminbi)
    CNY = Currency{code: "CNY", exponent: 2, name: "Chinese Yuan"}
    
    // INR - Indian Rupee
    // Note: Actually uses 2 decimal places in practice, not 3 as ISO specifies
    INR = Currency{code: "INR", exponent: 2, name: "Indian Rupee"}
    
    // BRL - Brazilian Real
    BRL = Currency{code: "BRL", exponent: 2, name: "Brazilian Real"}
    
    // KRW - South Korean Won
    // Note: Actually 0 decimal places in practice
    KRW = Currency{code: "KRW", exponent: 0, name: "South Korean Won"}
    
    // MXN - Mexican Peso
    MXN = Currency{code: "MXN", exponent: 2, name: "Mexican Peso"}
    
    // SGD - Singapore Dollar
    SGD = Currency{code: "SGD", exponent: 2, name: "Singapore Dollar"}
    
    // HKD - Hong Kong Dollar
    HKD = Currency{code: "HKD", exponent: 2, name: "Hong Kong Dollar"}
    
    // NZD - New Zealand Dollar
    NZD = Currency{code: "NZD", exponent: 2, name: "New Zealand Dollar"}
    
    // SEK - Swedish Krona
    SEK = Currency{code: "SEK", exponent: 2, name: "Swedish Krona"}
    
    // NOK - Norwegian Krone
    NOK = Currency{code: "NOK", exponent: 2, name: "Norwegian Krone"}
    
    // DKK - Danish Krone
    DKK = Currency{code: "DKK", exponent: 2, name: "Danish Krone"}
    
    // PLN - Polish Zloty
    PLN = Currency{code: "PLN", exponent: 2, name: "Polish Zloty"}
    
    // THB - Thai Baht
    THB = Currency{code: "THB", exponent: 2, name: "Thai Baht"}
    
    // ZAR - South African Rand
    ZAR = Currency{code: "ZAR", exponent: 2, name: "South African Rand"}
    
    // RUB - Russian Ruble
    RUB = Currency{code: "RUB", exponent: 2, name: "Russian Ruble"}
    
    // TRY - Turkish Lira
    TRY = Currency{code: "TRY", exponent: 2, name: "Turkish Lira"}
    
    // TWD - New Taiwan Dollar
    TWD = Currency{code: "TWD", exponent: 0, name: "New Taiwan Dollar"}
    
    // VND - Vietnamese Dong
    // Note: Actually 0 decimal places in practice
    VND = Currency{code: "VND", exponent: 0, name: "Vietnamese Dong"}
)
```

---

## Crypto Currencies (Optional)

```go
var (
    // BTC - Bitcoin
    // Note: Variable decimal places depending on context
    // Often represented with 8 decimal places (satoshi)
    BTC = Currency{
        code:     "BTC",
        exponent: 8,  // Satoshi precision
        name:     "Bitcoin",
    }
    
    // ETH - Ethereum
    // Note: Variable decimal places (wei precision)
    ETH = Currency{
        code:     "ETH",
        exponent: 18, // Wei precision
        name:     "Ethereum",
    }
)
```

**Warning**: Cryptocurrency handling requires special consideration:
- Variable precision requirements
- Different consensus rules
- Exchange-specific representations
- Not covered by ISO 4217

---

## ISO 4217 Code Validation

### Validation Function

```go
// ParseCurrency validates and returns a Currency for the given ISO 4217 code.
func ParseCurrency(code string) (Currency, error) {
    if len(code) != 3 {
        return Currency{}, ErrInvalidCurrencyCode
    }
    
    upper := strings.ToUpper(code)
    
    // Check cache first (avoid map lookup in hot path)
    if currency, ok := currencyByCode[upper]; ok {
        return currency, nil
    }
    
    return Currency{}, ErrUnknownCurrency
}

var currencyByCode = map[string]Currency{
    "USD": USD,
    "EUR": EUR,
    // ... all currencies
}
```

### Custom Currency Registration

```go
// RegisterCurrency adds a custom currency for the session/application.
// This is useful for corporate currencies, loyalty points, etc.
func RegisterCurrency(currency Currency) error {
    if len(currency.code) != 3 {
        return ErrInvalidCurrencyCode
    }
    
    currencyByCode[currency.code] = currency
    return nil
}

// Example: Corporate points system
func init() {
    RegisterCurrency(Currency{
        code:     "PTS",
        exponent: 0,
        name:     "Loyalty Points",
    })
}
```

---

## Currency Mismatch Detection

### Error Type

```go
// CurrencyMismatchError is returned when operations are attempted
// between Money values of different currencies.
type CurrencyMismatchError struct {
    Left     Currency  // Left-hand operand currency
    Right    Currency  // Right-hand operand currency
    Operation string   // Operation attempted (e.g., "add", "subtract")
}

func (e CurrencyMismatchError) Error() string {
    return fmt.Sprintf("currency mismatch: cannot %s %s (%d decimals) and %s (%d decimals)",
        e.Operation, e.Left.code, e.Left.exponent, e.Right.code, e.Right.exponent)
}
```

### Where Mismatch Detection Occurs

```go
// All binary operations check currency compatibility

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, CurrencyMismatchError{
            Left:      m.currency,
            Right:     other.currency,
            Operation: "add",
        }
    }
    // Proceed with addition
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, CurrencyMismatchError{
            Left:      m.currency,
            Right:     other.currency,
            Operation: "subtract",
        }
    }
    return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

func (m Money) Cmp(other Money) (int, error) {
    if m.currency != other.currency {
        return 0, CurrencyMismatchError{
            Left:      m.currency,
            Right:     other.currency,
            Operation: "compare",
        }
    }
    // Compare amounts
    switch {
    case m.amount < other.amount:
        return -1, nil
    case m.amount > other.amount:
        return 1, nil
    default:
        return 0, nil
    }
}
```

### Comparison Without Operation

```go
// AreSameCurrency checks if two currencies are equal without error.
func AreSameCurrency(a, b Currency) bool {
    return a == b
}

// CanOperate returns true if two currencies can be used in binary operations.
func CanOperate(a, b Currency) bool {
    return a == b
}
```

---

## Currency Conversion

### Explicit Conversion Required

Currency conversion requires an exchange rate and is never implicit:

```go
// Convert creates a new Money value in the target currency.
// Requires explicit exchange rate.
func (m Money) Convert(target Currency, rate Decimal, rounding RoundingMode) (Money, error) {
    if m.currency == target {
        return m, nil  // No conversion needed
    }
    
    // Convert: amount * rate
    converted := m.ToDecimal().Mul(rate)
    
    // Round to target currency's precision
    rounded := converted.Round(rounding, target.DecimalPlaces())
    
    return target.FromDecimal(rounded), nil
}
```

### Example Usage

```go
usd := USD.FromString("100.00")
rate := Decimal.New("1.08")  // EUR/USD rate

// Convert USD to EUR
eur, err := usd.Convert(EUR, rate, RoundHalfEven)
// Result: €92.59
```

---

## Scale/Precision Handling

### Scale from Currency

```go
// Scale returns the number of decimal places for this currency.
func (c Currency) Scale() int {
    return c.exponent
}

// ToUnscaled returns the integer representation.
// $19.99 with USD (scale 2) returns 1999
func (c Currency) ToUnscaled(money Money) int64 {
    if money.currency != c {
        panic("currency mismatch")
    }
    return money.amount
}
```

### Precision in Operations

```go
// Result precision rules:
// Add/Sub: Result has same scale as operands
// Mul: Scale = left.scale + right.scale
// Div: Scale = left.scale + additional precision (configurable, default 4)

// Multiply preserves precision
usd1 := USD.FromString("10.00")   // scale: 2
usd2 := USD.FromString("2.50")     // scale: 2
result := usd1.Mul(usd2)          // scale: 4 = $25.0000

// Division increases precision
usd := USD.FromString("10.00")
result := usd.Div(Decimal.NewInt(3)) // scale: 6 = $3.333333
```

---

## Zero Value Handling

```go
// Zero returns the zero value for this currency.
func (c Currency) Zero() Money {
    return Money{amount: 0, currency: c}
}

// IsZero returns true if the money value is zero.
func (m Money) IsZero() bool {
    return m.amount == 0
}

// IsZeroWithCurrency checks if money is zero in its currency context.
func (m Money) IsZeroWithCurrency() bool {
    return m.amount == 0 && m.currency != Currency{}
}
```

---

## Currency Comparison

```go
// EqualCurrency checks if two currencies are the same.
func (a Currency) EqualCurrency(b Currency) bool {
    return a.code == b.code
}

// Equal checks if two Money values are exactly equal (same currency and amount).
func (m Money) Equal(other Money) bool {
    return m.currency == other.currency && m.amount == other.amount
}
```

---

## Best Practices

### 1. Always Use Predefined Currencies When Possible

```go
// GOOD: Uses predefined currency
price := money.USD.FromString("19.99")

// AVOID: String-based currency (loses type safety)
price := NewMoney("19.99", "USD")
```

### 2. Use Currency Constants in Functions

```go
// GOOD: Function accepts Currency as parameter
func calculatePrice(currency money.Currency, basePrice int64) money.Money {
    return currency.FromInt(basePrice)
}

// ACCEPTABLE: Function specific to one currency
func calculateUSDPrice(basePrice int64) money.Money {
    return money.USD.FromInt(basePrice)
}

// AVOID: String-based currency in function signature
func calculatePrice(amount int64, currencyCode string) money.Money
```

### 3. Validate External Currency Codes

```go
// GOOD: Validate before use
func processPayment(currencyCode string, amount int64) error {
    currency, err := money.ParseCurrency(currencyCode)
    if err != nil {
        return fmt.Errorf("invalid currency: %w", err)
    }
    
    money := currency.FromInt(amount)
    // Process payment...
}
```

### 4. Handle Currency Mismatch Explicitly

```go
// GOOD: Check for currency mismatch errors
total, err := lineItemsTotal(items)
if err != nil {
    if _, ok := err.(money.CurrencyMismatchError); ok {
        log.Error().Err(err).Msg("Line items have mixed currencies")
        return ErrMixedCurrencies
    }
    return err
}
```

---

## Testing Currency Handling

### Currency Mismatch Tests

```go
func TestCurrencyMismatch(t *testing.T) {
    usd := money.USD.FromInt(1000)
    eur := money.EUR.FromInt(1000)
    
    _, err := usd.Add(eur)
    require.Error(t, err)
    
    var mismatch money.CurrencyMismatchError
    require.ErrorAs(t, err, &mismatch)
    require.Equal(t, "USD", mismatch.Left.Code())
    require.Equal(t, "EUR", mismatch.Right.Code())
}
```

### Scale Tests

```go
func TestCurrencyScale(t *testing.T) {
    require.Equal(t, 2, money.USD.Scale())
    require.Equal(t, 0, money.JPY.Scale())
    require.Equal(t, 8, money.BTC.Scale())
    
    // Test that values are stored correctly
    usd := money.USD.FromString("123.45")
    require.Equal(t, int64(12345), usd.Amount())
    
    jpy := money.JPY.FromString("500")
    require.Equal(t, int64(500), jpy.Amount())
}
```

---

## Summary

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Currency location | Type member | Compile-time safety |
| Validation | ISO 4217 + custom registration | Standard compliance + flexibility |
| Mismatch detection | Runtime error with context | Clear error messages |
| Conversion | Explicit with rate parameter | No silent conversions |
| Scale | From Currency definition | Reduces API complexity |
| Zero | Per-currency zero value | Type-safe zero |