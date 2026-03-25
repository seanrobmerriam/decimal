# ADR 0002: Default Rounding Mode

## Status
Accepted

## Date
2024-01-15

## Context
Financial calculations require consistent and predictable rounding behavior. When performing division or truncating operations that result in more precision than the display or storage format allows, the library must decide how to handle the remaining precision. This decision has significant implications for:

- Regulatory compliance (GAAP, IFRS)
- Financial statement accuracy
- Cumulative rounding error in repeated calculations
- Interoperability with other financial systems

### Rounding Mode Options Considered

1. **RoundHalfEven (Banker's Rounding)**: Round half to nearest even number
2. **RoundHalfUp**: Round half away from zero
3. **RoundHalfDown**: Round half toward zero
4. **RoundUp**: Always round away from zero
5. **RoundDown**: Always round toward zero (truncation)
6. **RoundCeiling**: Round toward positive infinity
7. **RoundFloor**: Round toward negative infinity

## Decision
We chose **RoundHalfEven (Banker's Rounding)** as the default rounding mode for the Decimal Money Library.

### Rationale

1. **IEEE 754 Compliance**: RoundHalfEven is the default rounding mode specified by IEEE 754 and is consistent with most hardware floating-point implementations.

2. **Reduces Systematic Bias**: In repeated calculations, RoundHalfUp introduces a systematic upward bias because it always rounds 0.5 upward. Banker's rounding distributes rounding decisions more evenly, reducing cumulative error.

3. **Financial Regulations**: Many financial regulations and accounting standards (GAAP, IFRS) are compatible with banker's rounding, and it is the de facto standard in banking systems.

4. **Predictable Behavior**: The rule "round half to even" is simple to understand and verify:
   - 1.5 rounds to 2 (even)
   - 2.5 rounds to 2 (even)
   - 3.5 rounds to 4 (even)

5. **Industry Practice**: Major financial libraries and programming languages use banker's rounding as their default:
   - Python's `Decimal.ROUND_HALF_EVEN`
   - Java's `BigDecimal.ROUND_HALF_EVEN`
   - .NET's `MidpointRounding.ToEven`

## Consequences

### Positive
- Consistent with IEEE 754 and most financial regulations
- Reduces systematic bias in repeated calculations
- Familiar to developers coming from other languages
- Predictable and verifiable behavior

### Negative
- May surprise developers expecting traditional "round half up"
- Requires explicit configuration for use cases requiring different rounding
- Slightly more complex to implement than simple truncation

### Mitigations
- Provide easy-to-use configuration for changing the default rounding mode
- Document the default clearly with examples
- Provide constants for all supported rounding modes
- Include examples showing how to use alternative rounding modes

## Alternatives Considered

### RoundHalfUp
- **Pros**: Intuitive, traditional rounding
- **Cons**: Introduces systematic upward bias in repeated calculations; not IEEE 754 default

### Truncation (RoundDown)
- **Pros**: Simple, deterministic, no rounding needed
- **Cons**: Loses information systematically; not suitable for financial reporting

## Implementation Notes

The library exposes rounding modes through the `RoundingMode` type:

```go
type RoundingMode int

const (
    RoundHalfEven RoundingMode = iota
    RoundHalfUp
    RoundHalfDown
    RoundUp
    RoundDown
    RoundCeiling
    RoundFloor
)
```

Users can configure the default rounding mode:

```go
// Set default for new Money values
money.SetDefaultRoundingMode(money.RoundHalfUp)

// Or specify per-operation
result := a.Add(b, money.RoundHalfUp)
```

## References

- IEEE 754-2019 Standard for Floating-Point Arithmetic
- GAAP Revenue Recognition Guidelines
- IFRS 15 Revenue from Contracts with Customers
- Python Decimal Documentation
- Java BigDecimal Documentation
