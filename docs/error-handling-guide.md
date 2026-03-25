# Error Handling Guide

This guide covers error handling patterns and best practices for the Decimal Money Library.

## Table of Contents

1. [Error Types Overview](#error-types-overview)
2. [Handling CurrencyMismatchError](#handling-currencymismatcherror)
3. [Handling DivisionByZero](#handling-divisionbyzero)
4. [Overflow Recovery with BigMoney](#overflow-recovery-with-bigmoney)
5. [Parsing Error Recovery](#parsing-error-recovery)
6. [Best Practices for User-Facing Errors](#best-practices-for-user-facing-errors)
7. [Go Error Handling Patterns](#go-error-handling-patterns)
8. [JavaScript Error Handling Patterns](#javascript-error-handling-patterns)

---

## Error Types Overview

The library defines several error types in the `money` package:

| Error Type | Code | Cause |
|------------|------|-------|
| `CurrencyMismatchError` | `CURRENCY_MISMATCH` | Operations between different currencies |
| `DivisionByZeroError` | `DIVISION_BY_ZERO` | Division by zero |
| `OverflowError` | `OVERFLOW` | Result exceeds int64 range |
| `UnderflowError` | `UNDERFLOW` | Value too small for precision |
| `PrecisionLossError` | `PRECISION_LOSS` | Lost digits in conversion |
| `ParseError` | `PARSE_ERROR` | String parsing failed |
| `InvalidOperationError` | `INVALID_OPERATION` | Invalid operation attempted |
| `RoundingError` | `ROUNDING_ERROR` | Rounding produced unexpected results |

All errors implement the `Error` interface with methods:
- `Kind() ErrorKind` - Category for type switching
- `Code() string` - Machine-readable error code
- `Details() map[string]interface{}` - Additional context

---

## Handling CurrencyMismatchError

`CurrencyMismatchError` occurs when you attempt arithmetic operations between Money values of different currencies.

### Problem

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    usd := money.USD.FromString("100.00")
    eur := money.EUR.FromString("100.00")
    
    result, err := usd.Add(eur)
    if err != nil {
        fmt.Println(err) // currency mismatch: cannot add USD (2 decimals) and EUR (2 decimals)
    }
}
```

### Solution 1: Type Assertion

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    usd := money.USD.FromString("100.00")
    eur := money.EUR.FromString("100.00")
    
    result, err := usd.Add(eur)
    if err != nil {
        if merr, ok := err.(money.CurrencyMismatchError); ok {
            fmt.Printf("Currency mismatch between %s and %s\n", 
                merr.Left.Code(), merr.Right.Code())
            fmt.Printf("Operation: %s\n", merr.Operation)
            // Handle the error appropriately
        }
    }
}
```

### Solution 2: Pre-Validation

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

// SafeAdd adds two Money values, returning an error if currencies don't match
func SafeAdd(a, b money.Money) (money.Money, error) {
    if a.Currency() != b.Currency() {
        return money.Money{}, fmt.Errorf(
            "cannot add %s and %s: currencies must match",
            a.Currency().Code(),
            b.Currency().Code(),
        )
    }
    return a.Add(b)
}

// EnsureCurrency converts Money to the target currency if possible
// For now, this is a no-op; in production you'd use exchange rates
func EnsureCurrency(m money.Money, target money.Currency) (money.Money, error) {
    if m.Currency() == target {
        return m, nil
    }
    return money.Money{}, fmt.Errorf(
        "currency conversion from %s to %s not supported",
        m.Currency().Code(),
        target.Code(),
    )
}

func main() {
    usd := money.USD.FromString("100.00")
    eur := money.EUR.FromString("100.00")
    
    // Pre-validate before operation
    if usd.Currency() == eur.Currency() {
        result, _ := usd.Add(eur)
        fmt.Println(result)
    } else {
        fmt.Println("Currencies don't match")
    }
}
```

---

## Handling DivisionByZero

`DivisionByZeroError` occurs when dividing by zero.

### Problem

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    amount := money.USD.FromString("100.00")
    zero := money.USD.FromString("0.00")
    
    result, err := amount.Div(zero.ToDecimal(), money.RoundHalfUp)
    if err != nil {
        fmt.Println(err) // division by zero
    }
}
```

### Solution 1: Pre-Check

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func safeDiv(dividend, divisor money.Money, rounding money.RoundingMode) (money.Money, error) {
    if divisor.IsZero() {
        return money.Money{}, fmt.Errorf("cannot divide by zero")
    }
    return dividend.Div(divisor.ToDecimal(), rounding)
}

func main() {
    amount := money.USD.FromString("100.00")
    zero := money.USD.FromString("0.00")
    
    result, err := safeDiv(amount, zero, money.RoundHalfUp)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Solution 2: Recoverable Result

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

// DivResult represents the result of a division operation
type DivResult struct {
    Quota money.Money
    Error error
}

// TryDivide attempts division, returning a result that always has a value or error
func TryDivide(dividend, divisor money.Money, rounding money.RoundingMode) DivResult {
    if divisor.IsZero() {
        // Return zero with error indicator
        return DivResult{
            Quota: dividend.Currency().Zero(),
            Error: fmt.Errorf("division by zero"),
        }
    }
    
    result, err := dividend.Div(divisor.ToDecimal(), rounding)
    return DivResult{Quota: result, Error: err}
}

func main() {
    amount := money.USD.FromString("100.00")
    zero := money.USD.FromString("0.00")
    
    result := TryDivide(amount, zero, money.RoundHalfUp)
    if result.Error != nil {
        fmt.Printf("Operation failed: %v\n", result.Error)
        fmt.Printf("Safe fallback value: %s\n", result.Quota)
    }
}
```

---

## Overflow Recovery with BigMoney

`OverflowError` occurs when a calculation result exceeds int64 range. Use `BigMoney` for arbitrary precision.

### Problem

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // This will overflow: national debt scale numbers
    debt := money.USD.FromInt(9223372036854775807) // Max int64
    
    // Try to double it - this will overflow
    doubled, err := debt.Multiply(money.USD.FromString("2"), money.RoundHalfUp)
    if err != nil {
        fmt.Printf("Overflow: %v\n", err)
    }
}
```

### Solution: BigMoney

```go
package main

import (
    "fmt"
    "github.com/decimal/money/bigmoney"
    "github.com/decimal/money/money"
)

func main() {
    // Use BigMoney for extreme scale
    debt, _ := bigmoney.NewBigMoneyFromString(money.USD, "9223372036854775807000")
    
    // Double it - no overflow with BigMoney
    doubled, err := debt.Mul(money.NewDecimalFromString("2"))
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Doubled: %s\n", doubled.String())
    }
    
    // Convert back to Money if it fits
    regularMoney := doubled.ToMoney()
    fmt.Printf("As regular Money: %s\n", regularMoney.String())
}
```

### Hybrid Approach

```go
package main

import (
    "fmt"
    "github.com/decimal/money/bigmoney"
    "github.com/decimal/money/money"
)

// TryMoney attempts an operation with Money, falls back to BigMoney on overflow
func TryMoney(a money.Money, op func(money.Money) (money.Money, error)) (money.Money, error) {
    result, err := op(a)
    if err != nil {
        // Check if it's an overflow error
        if merr, ok := err.(money.OverflowError); ok && merr.Code() == money.CodeOverflow {
            fmt.Println("Overflow detected, consider using BigMoney")
        }
        return money.Money{}, err
    }
    return result, nil
}
```

---

## Parsing Error Recovery

`ParseError` provides detailed information about string parsing failures.

### Problem

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    _, err := money.USD.FromString("$100.00")
    if err != nil {
        fmt.Println(err) // parse error: invalid character in integer part (expected digits only)
    }
}
```

### Solution: Detailed Error Handling

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func parseAmount(s string) (money.Money, error) {
    // Clean input
    cleaned := cleanMoneyString(s)
    
    money, err := money.USD.FromString(cleaned)
    if err != nil {
        if perr, ok := err.(money.ParseError); ok {
            return money.Money{}, fmt.Errorf(
                "cannot parse '%s' as money: %s (position %d, expected %s)",
                s,
                perr.Reason,
                perr.Position,
                perr.Expected,
            )
        }
        return money.Money{}, err
    }
    return money, nil
}

func cleanMoneyString(s string) string {
    // Remove currency symbols and whitespace
    s = strings.TrimSpace(s)
    s = strings.ReplaceAll(s, "$", "")
    s = strings.ReplaceAll(s, "€", "")
    s = strings.ReplaceAll(s, "£", "")
    s = strings.ReplaceAll(s, "¥", "")
    s = strings.ReplaceAll(s, ",", "")
    return s
}

func main() {
    testCases := []string{"$100.00", "€100.00", "100.00", "100", ""}
    
    for _, tc := range testCases {
        m, err := parseAmount(tc)
        if err != nil {
            fmt.Printf("'%s': Error - %v\n", tc, err)
        } else {
            fmt.Printf("'%s': %s\n", tc, m.String())
        }
    }
}
```

### Input Validation Pattern

```go
package main

import (
    "fmt"
    "regexp"
    "github.com/decimal/money/money"
)

var moneyPattern = regexp.MustCompile(`^-?[\d,]+\.?\d*$`)

func validateMoneyInput(s string) error {
    cleaned := strings.ReplaceAll(s, ",", "")
    if !moneyPattern.MatchString(cleaned) {
        return fmt.Errorf("invalid money format: %s", s)
    }
    return nil
}

func parseAmountValidated(s string) (money.Money, error) {
    if err := validateMoneyInput(s); err != nil {
        return money.Money{}, err
    }
    return money.USD.FromString(strings.ReplaceAll(s, ",", ""))
}
```

---

## Best Practices for User-Facing Errors

### Go: Custom Error Wrapping

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

// AppError represents an error with user-friendly messaging
type AppError struct {
    UserMessage string
    Err         error
}

func (e AppError) Error() string {
    return e.UserMessage
}

func (e AppError) Unwrap() error {
    return e.Err
}

// MoneyOperationError creates a user-friendly error for money operations
func MoneyOperationError(operation string, err error) AppError {
    switch err {
    case nil:
        return AppError{Err: nil}
    default:
        if _, ok := err.(money.CurrencyMismatchError); ok {
            return AppError{
                UserMessage: fmt.Sprintf("Cannot %s: amounts must be in the same currency", operation),
                Err:         err,
            }
        }
        if _, ok := err.(money.DivisionByZeroError); ok {
            return AppError{
                UserMessage: fmt.Sprintf("Cannot %s: cannot divide by zero", operation),
                Err:         err,
            }
        }
        if money.IsOverflow(err) {
            return AppError{
                UserMessage: fmt.Sprintf("Cannot %s: result is too large. Try breaking into smaller calculations.", operation),
                Err:         err,
            }
        }
        return AppError{
            UserMessage: fmt.Sprintf("Cannot %s: %v", operation, err),
            Err:         err,
        }
    }
}

func main() {
    usd := money.USD.FromString("100.00")
    eur := money.EUR.FromString("50.00")
    
    _, err := usd.Add(eur)
    appErr := MoneyOperationError("add amounts", err)
    
    fmt.Printf("User message: %s\n", appErr.UserMessage)
}
```

---

## Go Error Handling Patterns

### Pattern 1: Immediate Error Check

```go
result, err := operation()
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

### Pattern 2: Type Switch

```go
result, err := operation()
if err != nil {
    switch e := err.(type) {
    case money.CurrencyMismatchError:
        // Handle currency mismatch
        log.Printf("currency mismatch in %s: %s vs %s", 
            e.Operation, e.Left.Code(), e.Right.Code())
    case money.DivisionByZeroError:
        // Handle division by zero
        log.Printf("division by zero in %s", e.Context)
    case money.OverflowError:
        // Handle overflow
        log.Printf("overflow in %s: exceeds %d", e.Operation, e.MaxValue)
    default:
        return nil, fmt.Errorf("unknown error: %w", err)
    }
}
```

### Pattern 3: Error Codes

```go
result, err := operation()
if err != nil {
    if merr, ok := err.(money.Error); ok {
        switch merr.Code() {
        case money.CodeCurrencyMismatch:
            // Handle
        case money.CodeDivisionByZero:
            // Handle
        case money.CodeOverflow:
            // Handle
        }
    }
}
```

---

## JavaScript Error Handling Patterns

### Try-Catch with Error Messages

```javascript
import { Money, Currency } from '@decimal/money';

function calculateTotal(prices, taxRate) {
    try {
        let total = Money.fromString("USD", "0");
        
        for (const price of prices) {
            const amount = Money.fromString("USD", price);
            total = total.add(amount);
        }
        
        const tax = total.mul(Decimal.fromString(taxRate));
        return total.add(tax);
    } catch (err) {
        if (err.message.includes("currency mismatch")) {
            throw new Error("All prices must be in the same currency");
        }
        if (err.message.includes("division by zero")) {
            throw new Error("Tax rate cannot be zero");
        }
        throw err;
    }
}
```

### Validation Before Operation

```javascript
import { Money, Decimal } from '@decimal/money';

function safeDivide(a, b) {
    const divisor = typeof b === 'string' ? Decimal.fromString(b) : b;
    
    if (divisor.isZero()) {
        throw new Error("Division by zero: divisor cannot be zero");
    }
    
    return a.div(divisor);
}

function safeAdd(a, b) {
    if (a.currency.code !== b.currency.code) {
        throw new Error(
            `Currency mismatch: cannot add ${a.currency.code} and ${b.currency.code}`
        );
    }
    return a.add(b);
}
```

### Result Type Pattern

```javascript
// Either represents a value or an error, inspired by Rust's Result type
class Either {
    static left(value) {
        return { isLeft: true, value };
    }
    
    static right(error) {
        return { isRight: true, error };
    }
    
    static fromTryCatch(fn) {
        try {
            return Either.left(fn());
        } catch (e) {
            return Either.right(e);
        }
    }
}

function divideSafe(a, divisor) {
    return Either.fromTryCatch(() => {
        return a.div(divisor);
    });
}

const result = divideSafe(money, zeroDecimal);
if (result.isLeft) {
    console.log("Success:", result.value);
} else {
    console.log("Error:", result.error.message);
}
```

---

## See Also

- [Cookbook](cookbook.md) - Real-world usage patterns
- [Go API Design](go-api-design.md) - API design decisions
- [Testing Strategy](testing-strategy.md) - Testing error handling
