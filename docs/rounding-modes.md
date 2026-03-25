# Rounding Mode Specification

## Overview

Rounding is unavoidable in financial calculations when dividing values that don't divide evenly or when reducing precision. This document specifies all supported rounding modes, their precise definitions, use cases, and examples.

**Critical Note**: Rounding mode selection significantly impacts financial calculations. Incorrect rounding can lead to systematic gains or losses. Always align rounding mode with business requirements and regulatory guidance.

---

## Summary Table

| Mode | -1.5 | -1.4 | 1.4 | 1.5 | 1.15 | 1.25 |
|------|------|------|-----|-----|------|------|
| RoundUp | -2 | -2 | 2 | 2 | 2 | 2 |
| RoundDown | -1 | -1 | 1 | 1 | 1 | 1 |
| RoundHalfUp | -2 | -1 | 1 | 2 | 2 | 2 |
| RoundHalfDown | -1 | -1 | 1 | 1 | 1 | 2 |
| RoundHalfEven | -2 | -1 | 1 | 2 | 1 | 2 |
| RoundCeiling | -1 | -1 | 2 | 2 | 2 | 2 |
| RoundFloor | -2 | -2 | 1 | 1 | 1 | 1 |

*Examples rounded to nearest integer*

---

## Mode Definitions

### 1. RoundUp (Away from Zero)

**Definition**: Always round away from zero, regardless of the digit being dropped.

**Algorithm**:
```
if digit_to_drop > 0: increment <- +1
if digit_to_drop < 0: increment <- -1
round away from zero regardless
```

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.001 | $1.01 | Drop 1, increment up |
| $1.009 | $1.01 | Drop 9, increment up |
| $1.999 | $2.00 | Drop 9, increment up |
| $-1.001 | $-1.01 | Drop 1, increment away from zero |
| $-1.009 | $-1.01 | Drop 9, increment away from zero |

**Use Cases**:
- **Tax calculations** where the tax authority benefits
- **Conservative fee calculations** (always round up to customer's disadvantage)
- **Creating safety margins** in financial projections
- **Regulatory scenarios** where rounding against the party benefitting is required

**Financial Context**: Often used in consumer pricing ("always round up to your advantage")

**Tradeoffs**:
- Produces larger results, potentially systematic gain over many calculations
- Not suitable when fair distribution is required

---

### 2. RoundDown (Toward Zero)

**Definition**: Always round toward zero, simply truncating the excess digits.

**Algorithm**:
```
drop all digits beyond target precision
no increment regardless of dropped value
```

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.999 | $1.99 | Drop 9, no increment |
| $1.001 | $1.00 | Drop 0, no increment |
| $1.009 | $1.00 | Drop 9, no increment |
| $-1.999 | $-1.99 | Drop 9, no increment |
| $-1.001 | $-1.00 | Drop 1, no increment |

**Use Cases**:
- **Conservative estimates** where you want to understate values
- **Some tax jurisdictions** that require rounding down for tax calculations
- **Maximum discount calculations** (never give more than advertised)
- **When precision loss should always reduce value**

**Financial Context**: Common in promotional pricing where discounts are capped

**Tradeoffs**:
- Produces smaller results, systematic loss over many calculations
- Simple and predictable

---

### 3. RoundHalfUp (Round to Nearest, Tie Away from Zero)

**Definition**: When the digit being dropped is exactly 5 (or 5 followed by zeros), round away from zero. Otherwise, round to nearest neighbor.

**Algorithm**:
```
if dropped_digit >= 5: increment magnitude by 1
if dropped_digit < 5: no increment
if exactly 5 followed by zeros: increment away from zero
```

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.014 | $1.01 | Drop 4 < 5 |
| $1.015 | $1.02 | Drop 5, round up |
| $1.0150 | $1.02 | Exactly 5, round up |
| $1.0151 | $1.02 | Drop 1, round up |
| $1.019 | $1.02 | Drop 9 > 5 |
| $1.025 | $1.03 | Drop 5, round up |
| $-1.015 | $-1.02 | Drop 5, round away from zero |

**Use Cases**:
- **General-purpose rounding** for non-financial applications
- **Historical convention** in many programming languages
- **Human-readable results** (most people expect this behavior)

**Financial Context**: NOT recommended for financial calculations due to positive bias

**Tradeoffs**:
- Simple to understand and explain
- Introduces slight upward bias over many calculations
- Not suitable for regulated financial calculations

---

### 4. RoundHalfDown (Round to Nearest, Tie Toward Zero)

**Definition**: When the digit being dropped is exactly 5 (or 5 followed by zeros), round toward zero. Otherwise, round to nearest neighbor.

**Algorithm**:
```
if dropped_digit > 5: increment magnitude by 1
if dropped_digit < 5: no increment
if exactly 5 followed by zeros: round toward zero
```

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.014 | $1.01 | Drop 4 < 5 |
| $1.015 | $1.01 | Exactly 5, round down |
| $1.0150 | $1.01 | Exactly 5, round down |
| $1.0151 | $1.02 | Drop 1, round up |
| $1.025 | $1.02 | Exactly 5, round down |
| $1.0250 | $1.02 | Exactly 5, round down |
| $1.026 | $1.03 | Drop 6 > 5 |

**Use Cases**:
- **Reducing upward bias** compared to RoundHalfUp
- **Some European financial contexts**
- **When you want tie-breaking to favor smaller values**

**Financial Context**: Less common in financial systems

**Tradeoffs**:
- Introduces slight downward bias
- Asymmetric with RoundHalfUp

---

### 5. RoundHalfEven (Banker's Rounding)

**Definition**: When the digit being dropped is exactly 5 (or 5 followed by zeros), round to the nearest even number. Otherwise, round to nearest neighbor.

**Algorithm**:
```
if dropped_digit > 5: increment
if dropped_digit < 5: no increment
if exactly 5 followed by zeros:
    if last retained digit is even: no increment
    if last retained digit is odd: increment to make even
```

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.015 | $1.02 | 1 is odd, round up to 2 (even) |
| $1.025 | $1.02 | 2 is even, round down |
| $1.035 | $1.04 | 3 is odd, round up to 4 (even) |
| $1.045 | $1.04 | 4 is even, round down |
| $1.0150 | $1.02 | 1 is odd, round up |
| $1.01500 | $1.02 | 1 is odd, round up |
| $1.014999... | $1.01 | Slightly less than 5, round down |

**Mathematical Property**: This mode minimizes cumulative rounding error over large datasets because it distributes ties evenly.

**Use Cases**:
- **Financial calculations** requiring fairness over many operations
- **Regulatory compliance** in banking (recommended by IEEE 754)
- **Large batch calculations** where bias accumulation is a concern
- **Interbank settlements** and clearing house operations

**Financial Context**: **RECOMMENDED DEFAULT** for most financial calculations

**Regulatory Guidance**:
- **US**: Recommended by FDIC, OCC, Federal Reserve for currency calculations
- **EU**: Required by ECB for some operations
- **ISO 4217**: Doesn't mandate rounding mode, but RoundHalfEven is industry standard

**Tradeoffs**:
- Less intuitive for humans unfamiliar with banker's rounding
- Requires understanding for debugging
- Slightly more complex implementation

---

### 6. RoundCeiling (Toward +∞)

**Definition**: Always round toward positive infinity. If the value is positive or zero, round up; if negative, the magnitude increases toward zero (since -1.1 rounds to -1.0, which is "higher").

**Wait, Clarification on Sign**:
- RoundCeiling is always toward +∞
- For positive: 1.1 → 2.0, 1.01 → 2.0
- For negative: -1.1 → -1.0 (which IS toward +∞)

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.001 | $1.01 | Positive, round up |
| $1.999 | $2.00 | Positive, round up |
| $-1.001 | $-1.00 | Negative, round toward +∞ (less negative) |
| $-1.999 | $-1.99 | Negative, round toward +∞ |
| $0.00 | $0.00 | Zero unchanged |

**Use Cases**:
- **Profit calculations** where you want to ensure not understating gains
- **Minimum payment calculations**
- **Interest accrual** where borrower should benefit
- **Regulatory capital calculations** where conservative estimates required

**Financial Context**: Often used for consumer protection (ensuring minimum amounts are not understated)

**Tradeoffs**:
- Systematic upward bias
- Clear directional intent

---

### 7. RoundFloor (Toward -∞)

**Definition**: Always round toward negative infinity. If the value is negative or zero, round down; if positive, the magnitude decreases toward zero.

**Clarification on Sign**:
- RoundFloor is always toward -∞
- For negative: -1.1 → -2.0, -1.01 → -2.0
- For positive: 1.1 → 1.0 (which IS toward -∞)

**Examples** (rounding to 2 decimal places):

| Input | Result | Calculation |
|-------|--------|-------------|
| $1.001 | $1.00 | Positive, round down toward -∞ |
| $1.999 | $1.99 | Positive, round down |
| $-1.001 | $-1.01 | Negative, round down |
| $-1.999 | $-2.00 | Negative, round down |
| $0.00 | $0.00 | Zero unchanged |

**Use Cases**:
- **Liability calculations** where you want to ensure not overstating obligations
- **Maximum discount calculations** ensuring customer gets no more than entitled
- **Payment distribution** ensuring you don't overpay
- **Cost calculations** for expense recognition

**Financial Context**: Conservative approach for liabilities and costs

**Tradeoffs**:
- Systematic downward bias
- Clear directional intent

---

## Implementation Reference

### Go Implementation

```go
package decimal

// RoundingMode defines how to handle precision reduction.
type RoundingMode int

const (
    RoundUp RoundingMode = iota
    RoundDown
    RoundHalfUp
    RoundHalfDown
    RoundHalfEven
    RoundCeiling
    RoundFloor
)

// round applies the rounding mode to produce a value with targetScale.
// The value is represented as (unscaledInt, scale) where the actual value is
// unscaledInt / (10^scale).
func (mode RoundingMode) round(unscaledInt int64, scale, targetScale int) int64 {
    if scale <= targetScale {
        return unscaledInt
    }
    
    divisor := int64(pow10(scale - targetScale))
    remainder := unscaledInt % divisor
    quotient := unscaledInt / divisor
    
    if mode.needsIncrement(unscaledInt, divisor, remainder, quotient) {
        if unscaledInt < 0 {
            return quotient - 1
        }
        return quotient + 1
    }
    return quotient
}

func (mode RoundingMode) needsIncrement(unscaled int64, divisor, remainder, quotient int64) bool {
    drop := remainder * 2  // For tie detection
    absQuotient := quotient
    
    switch mode {
    case RoundUp:
        return remainder != 0 && unscaled >= 0 || (remainder != 0 && unscaled < 0 && mode == RoundUp)
        // Simplified: RoundUp increments if remainder != 0
        
    case RoundDown:
        return false
        
    case RoundHalfUp:
        return drop >= divisor
        
    case RoundHalfDown:
        return drop > divisor  // Strictly greater than, not equal
        
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
        return unscaled >= 0 && remainder != 0
        
    case RoundFloor:
        return unscaled < 0 && remainder != 0
    }
    return false
}
```

---

## Selection Guidelines

### When to Use RoundHalfEven (Banker's Rounding)

**DEFAULT CHOICE** for:
- Interest calculations over many periods
- Tax calculations across large datasets
- Any aggregation where bias accumulation matters
- Interchange and settlement calculations
- Any calculation that will be audited or reviewed by regulators

### When to Use RoundUp

- Tax on individual transactions
- Fee calculations (ensuring fees cover costs)
- Price rounding (always round price up)

### When to Use RoundDown

- Discount calculations (never exceed advertised discount)
- Conservative revenue projections

### When to Use RoundCeiling

- Minimum payment calculations
- Ensuring obligations are not understated
- Consumer protection scenarios

### When to Use RoundFloor

- Maximum discount enforcement
- Ensuring costs are not understated
- Liability calculations

---

## Common Mistakes

### Mistake 1: Implicit Rounding

**BAD**: Relying on language's default rounding (often RoundHalfUp)

```go
// BAD: Language may round 1.025 to 1.03 (RoundHalfUp)
price := 2.025
rounded := fmt.Sprintf("%.2f", price)  // "2.03" in some languages
```

**GOOD**: Explicit rounding mode

```go
// GOOD: Explicit rounding mode
result := money.USD.FromFloat(2.025).Round(RoundHalfEven, 2)
// Result: $2.02
```

### Mistake 2: Rounding at Every Step

**BAD**: Rounding intermediate results

```go
// BAD: Rounding on each operation accumulates error
subtotal := item.Div(3).Round(RoundHalfUp, 2)  // Round 1st
total := subtotal.Mul(3).Round(RoundHalfUp, 2)  // Round 2nd
// May not equal original!
```

**GOOD**: Preserve precision, round only at display/final output

```go
// GOOD: Maintain precision, round only at end
subtotals := []Money{item.Div(3), item.Div(3), item.Div(3)}
total := sum(subtotals).Round(RoundHalfEven, 2)
```

### Mistake 3: Mixing Rounding Modes

**BAD**: Different operations use different modes

```go
// BAD: Inconsistent rounding creates hard-to-audit calculations
discount := price.Mul(0.15).Round(RoundHalfUp, 2)   // Tax uses RoundUp
tax := discount.Mul(0.08).Round(RoundHalfDown, 2)   // Tax uses RoundDown
```

**GOOD**: Consistent rounding policy

```go
// GOOD: Consistent rounding mode throughout
DISCOUNT_ROUNDING := RoundHalfEven
TAX_ROUNDING := RoundHalfEven
```

---

## Regulatory Considerations

### United States

- **Sales Tax**: Many states specify rounding rules; RoundHalfUp is common
- **Banking**: OCC, FDIC, Federal Reserve recommend RoundHalfEven for currency calculations
- **SEC**: Investment calculations typically RoundHalfEven for fair value

### European Union

- **Euro Currency**: ECB specifies RoundHalfEven for interbank settlements
- **VAT**: Each member state may specify; most use RoundHalfUp per transaction

### Japan

- ** Yen Transactions**: No decimal rounding (0 decimal places)
- **Consumption Tax**: RoundHalfUp per transaction is common

### International Standards

- **IEEE 754**: Recommends RoundHalfEven as default for floating-point
- **ISO 4217**: Currency code standard, doesn't specify rounding
- **ISO 20022**: Financial messaging, allows implementation-defined rounding

---

## Summary

| Mode | Best For | Avoid When |
|------|----------|------------|
| RoundHalfEven | Default for financial calculations, bias minimization | When human-readable is priority |
| RoundUp | Conservative fees, tax (when party owes more) | Large datasets (bias accumulates) |
| RoundDown | Discounts, conservative estimates | When precision matters |
| RoundCeiling | Consumer protection, minimum amounts | Systematic upward bias is costly |
| RoundFloor | Liability protection, maximum constraints | Systematic downward bias is costly |
| RoundHalfUp | Simple human-readable results | Regulated financial calculations |
| RoundHalfDown | Reducing upward bias | Asymmetry concerns |