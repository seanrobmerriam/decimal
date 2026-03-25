# Decimal Money Library Cookbook

This cookbook provides practical patterns for common financial calculations using the Decimal Money Library.

## Table of Contents

1. [Tax Calculation](#tax-calculation)
2. [Interest Calculation](#interest-calculation)
3. [Allocation (Splitting Bills)](#allocation-splitting-bills)
4. [Currency Conversion](#currency-conversion)
5. [Double-Entry Accounting](#double-entry-accounting)
6. [Payroll Calculations](#payroll-calculations)
7. [Discount Calculations](#discount-calculations)
8. [Price Rounding](#price-rounding)

---

## Tax Calculation

### Sales Tax

Calculate sales tax on a purchase.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Calculate NYC sales tax (8.875%) on $29.99
    price := money.USD.FromString("29.99")
    taxRate := money.USD.FromString("0.08875")
    
    taxAmount, err := price.Multiply(taxRate, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    total := price.Add(taxAmount)
    fmt.Printf("Subtotal: %s\n", price.String())
    fmt.Printf("Tax (8.875%%): %s\n", taxAmount.String())
    fmt.Printf("Total: %s\n", total.String())
}
```

Output:
```
Subtotal: USD 29.99
Tax (8.875%): USD 2.66
Total: USD 32.65
```

### VAT Calculation

Calculate Value Added Tax with different rates.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // European VAT at 20%
    preTax := money.EUR.FromString("149.99")
    vatRate := money.EUR.FromString("0.20")
    
    vat, _ := preTax.Multiply(vatRate, money.RoundHalfUp)
    total, _ := preTax.Add(vat)
    
    fmt.Printf("Net: %s\n", preTax.String())
    fmt.Printf("VAT (20%%): %s\n", vat.String())
    fmt.Printf("Gross: %s\n", total.String())
}
```

### Multiple Tax Rates

Handle products with different tax rates.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func calculateTotalWithTaxes(items []struct {
    price money.Money
    rate  *money.Decimal
}) (money.Money, error) {
    var total money.Money
    
    for _, item := range items {
        tax, err := item.price.Multiply(item.rate, money.RoundHalfUp)
        if err != nil {
            return money.Money{}, err
        }
        taxed, err := item.price.Add(tax)
        if err != nil {
            return money.Money{}, err
        }
        total, err = total.Add(taxed)
        if err != nil {
            return money.Money{}, err
        }
    }
    
    return total, nil
}
```

---

## Interest Calculation

### Simple Interest

Calculate simple interest on a principal amount.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Calculate simple interest on $10,000 at 4.5% APR for 1 year
    principal := money.USD.FromInt(1000000) // $10,000.00 in cents
    rate := money.USD.FromString("0.045")     // 4.5% annual rate
    years := money.USD.FromInt(1)
    
    interest, err := principal.Multiply(rate, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    yearlyInterest, err := interest.Multiply(years, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    total, err := principal.Add(yearlyInterest)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Principal: %s\n", principal.String())
    fmt.Printf("Interest: %s\n", yearlyInterest.String())
    fmt.Printf("Total: %s\n", total.String())
}
```

### Monthly Interest (Amortization)

Calculate monthly interest for loan amortization.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Calculate monthly interest on $10,000 at 4.5% APR
    principal := money.USD.FromInt(1000000) // $10,000.00 in cents
    monthlyRate := money.USD.FromString("0.00375") // 4.5% / 12 months
    
    monthlyInterest, err := principal.Multiply(monthlyRate, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Monthly Interest: %s\n", monthlyInterest.String())
    // Output: Monthly Interest: USD 37.50
}
```

### Compound Interest

Calculate compound interest over multiple periods.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func compoundInterest(principal money.Money, rate *money.Decimal, periods int) (money.Money, error) {
    result := principal
    one := result.Currency().FromString("1.00")
    
    for i := 0; i < periods; i++ {
        interest, err := result.Multiply(rate, money.RoundHalfUp)
        if err != nil {
            return money.Money{}, err
        }
        periodTotal, err := result.Add(interest)
        if err != nil {
            return money.Money{}, err
        }
        result = periodTotal
        _ = one // suppress unused variable
    }
    
    return result, nil
}

func main() {
    // $10,000 at 5% compound annually for 10 years
    principal := money.USD.FromInt(1000000)
    rate := money.USD.FromString("0.05")
    
    futureValue, err := compoundInterest(principal, rate, 10)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Future Value: %s\n", futureValue.String())
    // Output: Future Value: USD 16288.95
}
```

---

## Allocation (Splitting Bills)

### Equal Split

Split a bill equally among multiple people.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/allocator"
    "github.com/decimal/money/money"
)

func main() {
    // Split $100 among 3 people
    total := money.USD.FromString("100.00")
    
    parts, err := allocator.Allocate(total, 3)
    if err != nil {
        panic(err)
    }
    
    for i, part := range parts {
        fmt.Printf("Person %d: %s\n", i+1, part.String())
    }
    // Output:
    // Person 1: USD 33.34
    // Person 2: USD 33.33
    // Person 3: USD 33.33
}
```

### Proportional Split

Split according to ratios (e.g., 5:3:2).

```go
package main

import (
    "fmt"
    "github.com/decimal/money/allocator"
    "github.com/decimal/money/money"
)

func main() {
    // Split $100 among ratios 5:3:2
    total := money.USD.FromString("100.00")
    
    parts, err := allocator.AllocateRatios(total, []int{5, 3, 2})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Alice (5 parts): %s\n", parts[0].String()) // $50.00
    fmt.Printf("Bob (3 parts): %s\n", parts[1].String())   // $30.00
    fmt.Printf("Carol (2 parts): %s\n", parts[2].String()) // $20.00
}
```

### Percentage-Based Split

Split according to percentages.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/allocator"
    "github.com/decimal/money/money"
)

func main() {
    // Split $1000 bonus: 50% to developer, 30% to designer, 20% to PM
    bonus := money.USD.FromString("1000.00")
    
    parts, err := allocator.AllocateByPercentages(bonus, []int{50, 30, 20})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Developer: %s\n", parts[0].String()) // $500.00
    fmt.Printf("Designer: %s\n", parts[1].String())  // $300.00
    fmt.Printf("PM: %s\n", parts[2].String())        // $200.00
}
```

---

## Currency Conversion

### Simple Conversion

Convert between currencies using a fixed exchange rate.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func convert(amount money.Money, targetCurrency money.Currency, rate *money.Decimal) (money.Money, error) {
    // Convert to Decimal first
    amountDecimal := &money.Decimal{
        Value: big.NewInt(amount.Amount()),
        Scale: int32(amount.Currency().DecimalPlaces()),
    }
    
    // Multiply by exchange rate
    converted, err := amountDecimal.Div(rate, money.RoundHalfUp)
    if err != nil {
        return money.Money{}, err
    }
    
    return targetCurrency.FromDecimal(converted), nil
}

func main() {
    // Convert USD to EUR at rate 0.85
    usdAmount := money.USD.FromString("100.00")
    eurRate := money.EUR.FromString("0.85")
    
    eurAmount, err := convert(usdAmount, money.EUR, &money.Decimal{})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("USD: %s\n", usdAmount.String())
    fmt.Printf("EUR (0.85): %s\n", eurAmount.String())
}
```

Note: In production, you would integrate with a real exchange rate service.

---

## Double-Entry Accounting

### Balanced Journal Entry

Ensure debits equal credits in a journal entry.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

// JournalEntry represents a double-entry accounting transaction
type JournalEntry struct {
    Description string
    Debits     []money.Money
    Credits    []money.Money
}

// IsBalanced checks if total debits equal total credits
func (e JournalEntry) IsBalanced() bool {
    var totalDebits, totalCredits money.Money
    
    for _, debit := range e.Debits {
        if totalDebits.IsZero() {
            totalDebits = debit
        } else {
            totalDebits, _ = totalDebits.Add(debit)
        }
    }
    
    for _, credit := range e.Credits {
        if totalCredits.IsZero() {
            totalCredits = credit
        } else {
            totalCredits, _ = totalCredits.Add(credit)
        }
    }
    
    return totalDebits.String() == totalCredits.String()
}

func main() {
    // Record a sale: Cash debited, Revenue credited
    entry := JournalEntry{
        Description: "Sale of goods",
        Debits: []money.Money{
            money.USD.FromString("100.00"),  // Cash
        },
        Credits: []money.Money{
            money.USD.FromString("100.00"),  // Revenue
        },
    }
    
    fmt.Printf("Entry: %s\n", entry.Description)
    fmt.Printf("Balanced: %v\n", entry.IsBalanced())
}
```

### Multiple Line Items

Handle journal entries with multiple debits and credits.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Purchase equipment with cash and loan
    // Equipment: $10,000 (debit)
    // Cash: $4,000 (credit)
    // Loan Payable: $6,000 (credit)
    
    equipment := money.USD.FromString("10000.00")
    cash := money.USD.FromString("4000.00")
    loan := money.USD.FromString("6000.00")
    
    // Verify debits = credits
    totalDebits := equipment
    totalCredits, _ := cash.Add(loan)
    
    isBalanced := totalDebits.String() == totalCredits.String()
    
    fmt.Printf("Total Debits: %s\n", totalDebits.String())
    fmt.Printf("Total Credits: %s\n", totalCredits.String())
    fmt.Printf("Balanced: %v\n", isBalanced)
}
```

---

## Payroll Calculations

### Hourly Wage Calculation

Calculate pay from hourly rate and hours worked.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Calculate bi-weekly pay for hourly employee
    hourlyRate := money.USD.FromString("25.50")
    hoursWorked := money.USD.FromString("80.00") // 40 hrs/week x 2 weeks
    
    grossPay, err := hourlyRate.Multiply(hoursWorked, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Gross Pay: %s\n", grossPay.String())
    // Output: Gross Pay: USD 2040.00
}
```

### Deduction Calculation

Calculate take-home pay after deductions.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    grossPay := money.USD.FromString("2040.00")
    
    // Calculate deductions
    federalTax, _ := grossPay.Multiply(money.USD.FromString("0.15"), money.RoundHalfUp)
    stateTax, _ := grossPay.Multiply(money.USD.FromString("0.05"), money.RoundHalfUp)
    socialSecurity, _ := grossPay.Multiply(money.USD.FromString("0.062"), money.RoundHalfUp)
    medicare, _ := grossPay.Multiply(money.USD.FromString("0.0145"), money.RoundHalfUp)
    
    var totalDeductions money.Money
    totalDeductions, _ = totalDeductions.Add(federalTax)
    totalDeductions, _ = totalDeductions.Add(stateTax)
    totalDeductions, _ = totalDeductions.Add(socialSecurity)
    totalDeductions, _ = totalDeductions.Add(medicare)
    
    netPay, _ := grossPay.Sub(totalDeductions)
    
    fmt.Printf("Gross Pay: %s\n", grossPay.String())
    fmt.Printf("Federal Tax: %s\n", federalTax.String())
    fmt.Printf("State Tax: %s\n", stateTax.String())
    fmt.Printf("Social Security: %s\n", socialSecurity.String())
    fmt.Printf("Medicare: %s\n", medicare.String())
    fmt.Printf("Total Deductions: %s\n", totalDeductions.String())
    fmt.Printf("Net Pay: %s\n", netPay.String())
}
```

---

## Discount Calculations

### Percentage Discount

Apply a percentage discount to a price.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    originalPrice := money.USD.FromString("99.99")
    discountPercent := money.USD.FromString("0.20") // 20% off
    
    discount, err := originalPrice.Multiply(discountPercent, money.RoundHalfUp)
    if err != nil {
        panic(err)
    }
    
    salePrice, err := originalPrice.Sub(discount)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Original: %s\n", originalPrice.String())
    fmt.Printf("Discount: %s\n", discount.String())
    fmt.Printf("Sale Price: %s\n", salePrice.String())
}
```

### Tiered Pricing

Apply different discounts based on quantity.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func applyTieredPricing(unitPrice money.Money, quantity int) (money.Money, error) {
    subtotal, err := unitPrice.Multiply(unitPrice.Currency().FromInt(int64(quantity)), money.RoundHalfUp)
    if err != nil {
        return money.Money{}, err
    }
    
    var discountRate *money.Decimal
    switch {
    case quantity >= 100:
        discountRate = money.USD.FromString("0.20") // 20% off
    case quantity >= 50:
        discountRate = money.USD.FromString("0.10") // 10% off
    case quantity >= 10:
        discountRate = money.USD.FromString("0.05") // 5% off
    default:
        discountRate = money.USD.FromString("0.00") // No discount
    }
    
    discount, err := subtotal.Multiply(discountRate, money.RoundHalfUp)
    if err != nil {
        return money.Money{}, err
    }
    
    return subtotal.Sub(discount)
}

func main() {
    unitPrice := money.USD.FromString("25.00")
    
    for _, qty := range []int{5, 15, 55, 105} {
        total, err := applyTieredPricing(unitPrice, qty)
        if err != nil {
            panic(err)
        }
        fmt.Printf("Quantity %d: %s\n", qty, total.String())
    }
}
```

---

## Price Rounding

### Half-Up Rounding

Standard rounding where 0.5 rounds up.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Round to nearest cent
    values := []string{"1.234", "1.235", "1.236"}
    
    for _, v := range values {
        d := money.NewDecimalFromString(v)
        rounded := d.Round(money.RoundHalfUp, 2)
        fmt.Printf("%s -> %s\n", v, rounded.String())
    }
    // Output:
    // 1.234 -> 1.23
    // 1.235 -> 1.24
    // 1.236 -> 1.24
}
```

### Banker's Rounding

Round to nearest even number on ties (used in accounting).

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // Compare HalfUp vs HalfEven
    d := money.NewDecimalFromString("1.025")
    
    halfUp := d.Round(money.RoundHalfUp, 2)
    halfEven := d.Round(money.RoundHalfEven, 2)
    
    fmt.Printf("1.025 with RoundHalfUp: %s\n", halfUp.String())   // 1.03
    fmt.Printf("1.025 with RoundHalfEven: %s\n", halfEven.String()) // 1.02
}
```

### Currency-Specific Rounding

Ensure prices match currency decimal places.

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func main() {
    // USD: 2 decimal places
    // JPY: 0 decimal places (no cents)
    
    usdAmount := money.USD.FromString("99.999")
    jpyAmount := money.JPY.FromString("99.999")
    
    fmt.Printf("USD %s (2 decimals): %s\n", "99.999", usdAmount.String())
    fmt.Printf("JPY %s (0 decimals): %s\n", "99.999", jpyAmount.String())
    // Output:
    // USD 99.999 (2 decimals): USD 100.00
    // JPY 99.999 (0 decimals): JPY 100
}
```

---

## Error Handling Patterns

### Handling Currency Mismatch

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func safeAdd(a, b money.Money) (money.Money, error) {
    if a.Currency() != b.Currency() {
        return money.Money{}, fmt.Errorf(
            "cannot add %s and %s: currency mismatch",
            a.Currency().Code(),
            b.Currency().Code(),
        )
    }
    return a.Add(b)
}

func main() {
    usd := money.USD.FromString("100.00")
    eur := money.EUR.FromString("100.00")
    
    result, err := safeAdd(usd, eur)
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

### Handling Division by Zero

```go
package main

import (
    "fmt"
    "github.com/decimal/money/money"
)

func safeDivide(dividend, divisor money.Money, rounding money.RoundingMode) (money.Money, error) {
    if divisor.IsZero() {
        return money.Money{}, fmt.Errorf("division by zero")
    }
    return dividend.Div(divisor, rounding)
}

func main() {
    amount := money.USD.FromString("100.00")
    zero := money.USD.FromString("0.00")
    
    result, err := safeDivide(amount, zero, money.RoundHalfUp)
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

---

## See Also

- [Error Handling Guide](error-handling-guide.md) - Detailed error handling patterns
- [Go API Design](../docs/go-api-design.md) - API design decisions
- [Rounding Modes](../docs/rounding-modes.md) - When to use each rounding mode
