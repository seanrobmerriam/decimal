# Error Handling Design

## Overview

Robust error handling is critical for a financial library. This document defines the error types, handling strategies, and recovery patterns for the decimal money library.

---

## Design Principles

1. **Errors are values**: Go-style error handling with explicit error returns
2. **Specific error types**: Concrete error types for each failure mode enable precise handling
3. **Rich context**: Errors include all relevant context for debugging
4. **No silent failures**: Operations that can fail return errors; no default-to-zero behavior
5. **Recoverable by default**: Use errors rather than panics for recoverable conditions

---

## Error Type Hierarchy

```
Error (interface/base)
├── CurrencyMismatchError
├── DivisionByZeroError  
├── OverflowError
├── UnderflowError
├── PrecisionLossError
├── ParseError
├── InvalidOperationError
└── RoundingError
```

---

## Base Error Interface

### Go Interface

```go
// Error represents an error in money operations.
type Error interface {
    error
    // Kind returns the error category.
    Kind() ErrorKind
    // Code returns a machine-readable error code.
    Code() string
    // Details returns additional error context.
    Details() map[string]interface{}
}

// ErrorKind categorizes errors for programmatic handling.
type ErrorKind int

const (
    ErrKindCurrencyMismatch ErrorKind = iota
    ErrKindDivisionByZero
    ErrKindOverflow
    ErrKindUnderflow
    ErrKindPrecisionLoss
    ErrKindParse
    ErrKindInvalidOperation
    ErrKindRounding
)
```

### Error Code Constants

```go
const (
    CodeCurrencyMismatch = "CURRENCY_MISMATCH"
    CodeDivisionByZero   = "DIVISION_BY_ZERO"
    CodeOverflow         = "OVERFLOW"
    CodeUnderflow        = "UNDERFLOW"
    CodePrecisionLoss    = "PRECISION_LOSS"
    CodeParseError       = "PARSE_ERROR"
    CodeInvalidOperation = "INVALID_OPERATION"
    CodeRounding         = "ROUNDING_ERROR"
)
```

---

## Concrete Error Types

### 1. CurrencyMismatchError

**When**: Operations between Money values of different currencies

```go
// CurrencyMismatchError is returned when operations are attempted
// between Money values of different currencies.
type CurrencyMismatchError struct {
    Left      Currency  // Left operand currency
    Right     Currency  // Right operand currency
    Operation string    // Operation name: "add", "sub", "mul", "div", "cmp"
}

func (e CurrencyMismatchError) Error() string {
    return fmt.Sprintf("currency mismatch: cannot %s %s (%d decimals) and %s (%d decimals)",
        e.Operation, e.Left.code, e.Left.exponent, e.Right.code, e.Right.exponent)
}

func (e CurrencyMismatchError) Kind() ErrorKind { return ErrKindCurrencyMismatch }
func (e CurrencyMismatchError) Code() string     { return CodeCurrencyMismatch }
func (e CurrencyMismatchError) Details() map[string]interface{} {
    return map[string]interface{}{
        "left_currency":  e.Left.code,
        "right_currency": e.Right.code,
        "operation":       e.Operation,
    }
}
```

**Example**:
```go
usd := USD.FromInt(1000)  // $10.00
eur := EUR.FromInt(1000)  // €10.00

_, err := usd.Add(eur)
// err: "currency mismatch: cannot add USD (2 decimals) and EUR (2 decimals)"
```

**Recovery Strategy**:
- Always check currency before operations
- Use `CanOperate(a, b Currency)` to pre-check
- Convert to common currency first

---

### 2. DivisionByZeroError

**When**: Division by zero (scalar or Money with zero amount)

```go
// DivisionByZeroError is returned when dividing by zero.
type DivisionByZeroError struct {
    Dividend   Money     // The value being divided
    Divisor    interface{} // The zero divisor (number or Money)
    Context    string     // Where the error occurred
}

func (e DivisionByZeroError) Error() string {
    return fmt.Sprintf("division by zero: cannot divide %s by zero", e.Dividend.String())
}

func (e DivisionByZeroError) Kind() ErrorKind { return ErrKindDivisionByZero }
func (e DivisionByZeroError) Code() string     { return CodeDivisionByZero }
func (e DivisionByZeroError) Details() map[string]interface{} {
    return map[string]interface{}{
        "dividend": e.Dividend.String(),
        "context":  e.Context,
    }
}
```

**Example**:
```go
usd := USD.FromInt(1000)

_, err := usd.Div(0, RoundHalfEven)
// err: "division by zero: cannot divide $10.00 by zero"
```

**Recovery Strategy**:
- Pre-check divisor is non-zero: `divisor.IsZero()`
- Handle zero-result case explicitly if mathematically valid
- Return infinity or error based on business requirements

---

### 3. OverflowError

**When**: Result exceeds maximum representable value

```go
// OverflowError is returned when a calculation result exceeds int64 range.
type OverflowError struct {
    Operation string                 // Operation that caused overflow
    Left      interface{}            // Left operand
    Right     interface{}            // Right operand  
    MaxValue  int64                  // Maximum representable value
    Context   map[string]interface{} // Additional context
}

func (e OverflowError) Error() string {
    return fmt.Sprintf("overflow: result exceeds maximum value %d from operation %s",
        e.MaxValue, e.Operation)
}

func (e OverflowError) Kind() ErrorKind { return ErrKindOverflow }
func (e OverflowError) Code() string     { return CodeOverflow }
func (e OverflowError) Details() map[string]interface{} {
    return e.Context
}
```

**Example**:
```go
max := USD.FromInt(math.MaxInt64)  // $9,223,372,036,854,775,807
huge := USD.FromInt(1000)

_, err := max.Mul(huge)
// err: "overflow: result exceeds maximum value 9223372036854775807"
```

**Recovery Strategy**:
- Use `BigMoney` for values exceeding int64
- Check magnitude before operations: `CanMultiply(m, f)` 
- Consider using saturating arithmetic (return max value)

---

### 4. UnderflowError

**When**: Result is too close to zero to represent

```go
// UnderflowError is returned when a calculation produces a value
// too small to represent at the required precision.
type UnderflowError struct {
    Operation    string  // Operation that caused underflow
    Value        Decimal // The value that underflowed
    MinPrecision int     // Minimum representable precision
    Context      string  // Where the error occurred
}

func (e UnderflowError) Error() string {
    return fmt.Sprintf("underflow: value %s below minimum representable at precision %d",
        e.Value.String(), e.MinPrecision)
}

func (e UnderflowError) Kind() ErrorKind { return ErrKindUnderflow }
func (e UnderflowError) Code() string     { return CodeUnderflow }
```

**Recovery Strategy**:
- Round to zero if value is below epsilon
- Increase working precision
- Return zero with optional warning

---

### 5. PrecisionLossError

**When**: Converting from float or losing precision in operation

```go
// PrecisionLossError is returned when precision is lost in a conversion or operation.
type PrecisionLossError struct {
    Original    string  // Original string representation
    Converted   string  // Converted value (approximation)
    LostDigits  int     // Number of significant digits lost
    Context     string  // Where the loss occurred
}

func (e PrecisionLossError) Error() string {
    return fmt.Sprintf("precision loss: %s converted to %s, lost %d digits",
        e.Original, e.Converted, e.LostDigits)
}

func (e PrecisionLossError) Kind() ErrorKind   { return ErrKindPrecisionLoss }
func (e PrecisionLossError) Code() string       { return CodePrecisionLoss }
```

**Example**:
```go
// WARNING: Precision loss from float
usd, err := USD.FromFloat(0.1 + 0.2)  
// usd: $0.30 (but actual value is $0.30000000000000004)
// err: "precision loss: 0.30000000000000004 converted to 0.30, lost digits"
```

**Recovery Strategy**:
- Use `FromString` or `FromInt` for exact values
- Accept precision loss with explicit acknowledgment
- Reject float conversion entirely in strict mode

---

### 6. ParseError

**When**: String cannot be parsed as Money or Decimal

```go
// ParseError is returned when a string cannot be parsed.
type ParseError struct {
    Input    string  // The string that failed to parse
    Position int     // Position of error (0 if unknown)
    Reason   string  // Why parsing failed
    Expected string  // What was expected
}

func (e ParseError) Error() string {
    if e.Position > 0 {
        return fmt.Sprintf("parse error at position %d: %s (expected %s)",
            e.Position, e.Reason, e.Expected)
    }
    return fmt.Sprintf("parse error: %s (expected %s)", e.Reason, e.Expected)
}

func (e ParseError) Kind() ErrorKind { return ErrKindParse }
func (e ParseError) Code() string     { return CodeParseError }
func (e ParseError) Details() map[string]interface{} {
    return map[string]interface{}{
        "input":    e.Input,
        "position": e.Position,
        "reason":   e.Reason,
        "expected": e.Expected,
    }
}
```

**Example**:
```go
_, err := USD.FromString("abc")
// err: "parse error: invalid character 'a' (expected number)"

_, err := USD.FromString("$10.99")
// err: "parse error: currency symbol not allowed (expected bare number)"
```

**Recovery Strategy**:
- Validate input format before parsing
- Strip currency symbols with `ParseCurrencyString`
- Use regex to pre-validate

---

### 7. InvalidOperationError

**When**: Operation not supported or invalid combination

```go
// InvalidOperationError is returned when an operation is not valid.
type InvalidOperationError struct {
    Operation string                 // The invalid operation
    Reason    string                 // Why it's invalid
    Operands  []interface{}          // The operands involved
    Context   map[string]interface{} // Additional context
}

func (e InvalidOperationError) Error() string {
    return fmt.Sprintf("invalid operation: %s - %s", e.Operation, e.Reason)
}

func (e InvalidOperationError) Kind() ErrorKind   { return ErrKindInvalidOperation }
func (e InvalidOperationError) Code() string       { return CodeInvalidOperation }
```

**Example**:
```go
// Cannot take percentage of negative money in some business rules
debt := USD.FromInt(-1000)  // -$10.00
_, err := debt.Percentage(10)  // Invalid? Depends on business rules
```

---

### 8. RoundingError

**When**: Rounding produces unexpected or inconsistent results

```go
// RoundingError is returned when rounding behavior is questionable.
type RoundingError struct {
    Original   Decimal  // Value before rounding
    Rounded    Decimal  // Value after rounding
    Mode       RoundingMode // Rounding mode used
    Reason     string   // Why rounding was problematic
}

func (e RoundingError) Error() string {
    return fmt.Sprintf("rounding warning: %s -> %s using %s (%s)",
        e.Original.String(), e.Rounded.String(), e.Mode.String(), e.Reason)
}

func (e RoundingError) Kind() ErrorKind { return ErrKindRounding }
func (e RoundingError) Code() string     { return CodeRounding }
```

**Recovery Strategy**:
- RoundingError is a warning; operation succeeded
- Log for audit if required
- Consider using different rounding mode

---

## Error Handling Patterns

### Pattern 1: Early Validation

```go
func processPayment(amount Money, account Account) error {
    // Validate currencies match
    if amount.Currency() != account.Balance().Currency() {
        return CurrencyMismatchError{
            Left:      amount.Currency(),
            Right:     account.Balance().Currency(),
            Operation: "process payment",
        }
    }
    
    // Validate sufficient balance
    if amount.GreaterThan(account.Balance()) {
        return InsufficientFundsError{
            Required: amount,
            Available: account.Balance(),
        }
    }
    
    // Proceed with payment
    return nil
}
```

### Pattern 2: Error Aggregation

```go
// Process multiple items, collect all errors
func reconcileItems(items []Money, expected Money) (errs []error) {
    var sum Money
    for i, item := range items {
        if s, err := sum.Add(item); err != nil {
            errs = append(errs, fmt.Errorf("item %d: %w", i, err))
        } else {
            sum = s
        }
    }
    
    if !sum.Equal(expected) {
        errs = append(errs, ReconciliationError{
            Expected: expected,
            Actual:   sum,
            Diff:     func() Money { 
                diff, _ := sum.Sub(expected)
                return diff
            }(),
        })
    }
    return
}
```

### Pattern 3: Try-Helper Pattern

```go
// Must succeeds or panics - useful for tests
func (m Money) MustAdd(other Money) Money {
    result, err := m.Add(other)
    if err != nil {
        panic(err)
    }
    return result
}

// Try returns zero value on error - for non-critical paths
func (m Money) TryAdd(other Money) Money {
    result, err := m.Add(other)
    if err != nil {
        return Money{} // Zero value
    }
    return result
}
```

---

## Global Error Configuration

### Panic Recovery

```go
// Configure whether certain errors should panic (for development)
var DebugMode = false

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        err := CurrencyMismatchError{...}
        if DebugMode {
            panic(err)
        }
        return Money{}, err
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
```

### Custom Error Handlers

```go
// ErrorHandler allows customization of error behavior
type ErrorHandler interface {
    HandleError(Error) error  // Transform or log error
}

// DefaultHandler returns errors as-is
type DefaultErrorHandler struct{}

func (h DefaultErrorHandler) HandleError(err Error) error {
    return err
}

// LoggingHandler logs errors before returning
type LoggingErrorHandler struct {
    Logger *log.Logger
}

func (h LoggingErrorHandler) HandleError(err Error) error {
    h.Logger.Printf("decimal error: %v", err)
    return err
}
```

---

## Wasm Error Handling

### JavaScript Error Objects

```rust
#[wasm_bindgen]
pub struct WasmError {
    kind: String,
    message: String,
    code: String,
}

#[wasm_bindgen]
impl WasmError {
    #[wasm_bindgen(getter)]
    pub fn kind(&self) -> String {
        self.kind.clone()
    }
    
    #[wasm_bindgen(getter)]
    pub fn message(&self) -> String {
        self.message.clone()
    }
    
    #[wasm_bindgen(getter)]
    pub fn code(&self) -> String {
        self.code.clone()
    }
}

impl From<Error> for WasmError {
    fn from(err: Error) -> WasmError {
        WasmError {
            kind: format!("{:?}", err.kind()),
            code: err.code(),
            message: err.to_string(),
        }
    }
}
```

### JavaScript Usage

```javascript
try {
    const result = money.add(other);
} catch (e) {
    if (e.code === 'CURRENCY_MISMATCH') {
        console.error('Cannot add different currencies');
    } else if (e.code === 'DIVISION_BY_ZERO') {
        console.error('Cannot divide by zero');
    }
}
```

---

## Error Logging and Auditing

### Audit Trail

```go
// WithAuditing wraps operations with audit logging
func WithAuditing(op string, a, b Money) (Money, error) {
    result, err := op(a, b)
    
    AuditLog.Printf("decimal operation: op=%s left=%s right=%s result=%s error=%v",
        op, a, b, result, err)
    
    return result, err
}
```

### Error Metrics

```go
var errorMetrics = metrics.NewCounter("decimal_errors_total")

func (e CurrencyMismatchError) Error() string {
    errorMetrics.WithLabelValues("currency_mismatch").Inc()
    return fmt.Sprintf(...)
}
```

---

## Testing Error Paths

```go
func TestCurrencyMismatchError(t *testing.T) {
    usd := USD.FromInt(1000)
    eur := EUR.FromInt(1000)
    
    _, err := usd.Add(eur)
    require.Error(t, err)
    
    // Check error type
    var mismatchErr CurrencyMismatchError
    require.ErrorAs(t, err, &mismatchErr)
    
    // Check details
    require.Equal(t, "USD", mismatchErr.Left.Code())
    require.Equal(t, "EUR", mismatchErr.Right.Code())
    require.Equal(t, "add", mismatchErr.Operation)
    
    // Check error string
    require.Contains(t, err.Error(), "currency mismatch")
    require.Contains(t, err.Error(), "USD")
    require.Contains(t, err.Error(), "EUR")
}

func TestDivisionByZero(t *testing.T) {
    usd := USD.FromInt(1000)
    
    _, err := usd.Div(0, RoundHalfEven)
    require.Error(t, err)
    
    var divZeroErr DivisionByZeroError
    require.ErrorAs(t, err, &divZeroErr)
    require.Equal(t, "$10.00", divZeroErr.Dividend.String())
}

func TestOverflow(t *testing.T) {
    max := USD.FromInt(math.MaxInt64)
    huge := USD.FromInt(2)
    
    _, err := max.Mul(huge)
    require.Error(t, err)
    
    var overflowErr OverflowError
    require.ErrorAs(t, err, &overflowErr)
    require.Contains(t, err.Error(), "overflow")
}
```

---

## Summary

| Error Type | When | Recovery |
|------------|------|----------|
| CurrencyMismatchError | Mixed currencies | Convert first |
| DivisionByZeroError | Divide by zero | Check divisor |
| OverflowError | Exceeds int64 | Use BigMoney |
| UnderflowError | Below minimum | Round to zero |
| PrecisionLossError | Float conversion | Use string/int |
| ParseError | Invalid string | Validate input |
| InvalidOperationError | Invalid operation | Fix operands |
| RoundingError | Rounding concern | Log and proceed |