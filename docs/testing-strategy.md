# Testing Strategy

## Overview

This document outlines the comprehensive testing strategy for the decimal money library, combining property-based testing, scenario-based testing, and edge case coverage to ensure correctness for financial applications.

---

## Testing Pyramid

```
                    ┌───────────────┐
                    │  Integration  │
                    │    Tests       │
                    │  (Wasm, APIs) │
                    ├───────────────┤
                    │   Scenario    │
                    │    Tests       │
                    │ (Financial)   │
                    ├───────────────┤
                    │   Property     │
                    │    Tests       │
                    │  (Invariants) │
                    ├───────────────┤
                    │     Unit       │
                    │    Tests       │
                    │  (Operations) │
                    └───────────────┘
```

---

## 1. Unit Tests

### Operation Tests

```go
package money_test

import (
    "testing"
    "github.com/decimal/money"
)

func TestMoneyAdd(t *testing.T) {
    cases := []struct {
        name     string
        a        money.Money
        b        money.Money
        expected money.Money
    }{
        {
            name:     "simple addition",
            a:        money.USD.FromInt(1000),  // $10.00
            b:        money.USD.FromInt(500),   // $5.00
            expected: money.USD.FromInt(1500),  // $15.00
        },
        {
            name:     "addition with carry",
            a:        money.USD.FromInt(999),   // $9.99
            b:        money.USD.FromInt(2),     // $0.02
            expected: money.USD.FromInt(1001),  // $10.01
        },
        {
            name:     "large values",
            a:        money.USD.FromInt(999999999999),  // $9,999,999.99
            b:        money.USD.FromInt(1),              // $0.01
            expected: money.USD.FromInt(1000000000000),  // $10,000,000.00
        },
        {
            name:     "negative plus positive",
            a:        money.USD.FromInt(-500),   // -$5.00
            b:        money.USD.FromInt(1000),   // $10.00
            expected: money.USD.FromInt(500),    // $5.00
        },
        {
            name:     "both negative",
            a:        money.USD.FromInt(-1000),  // -$10.00
            b:        money.USD.FromInt(-500),    // -$5.00
            expected: money.USD.FromInt(-1500),   // -$15.00
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tc.a.Add(tc.b)
            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}

func TestMoneySub(t *testing.T) {
    cases := []struct {
        name     string
        a        money.Money
        b        money.Money
        expected money.Money
    }{
        {
            name:     "simple subtraction",
            a:        money.USD.FromInt(1000),  // $10.00
            b:        money.USD.FromInt(300),    // $3.00
            expected: money.USD.FromInt(700),   // $7.00
        },
        {
            name:     "result negative",
            a:        money.USD.FromInt(100),   // $1.00
            b:        money.USD.FromInt(500),    // $5.00
            expected: money.USD.FromInt(-400),   // -$4.00
        },
        {
            name:     "subtract zero",
            a:        money.USD.FromInt(1000),   // $10.00
            b:        money.USD.FromInt(0),      // $0.00
            expected: money.USD.FromInt(1000),   // $10.00
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tc.a.Sub(tc.b)
            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}

func TestMoneyMul(t *testing.T) {
    cases := []struct {
        name     string
        money    money.Money
        scalar   interface{}
        expected money.Money
    }{
        {
            name:     "multiply by int",
            money:    money.USD.FromInt(1000),   // $10.00
            scalar:   int64(3),
            expected: money.USD.FromInt(3000),   // $30.00
        },
        {
            name:     "multiply by decimal",
            money:    money.USD.FromInt(1000),              // $10.00
            scalar:   money.Decimal.NewFloat(1.0775),       // tax rate
            expected: money.USD.FromInt(1078),              // $10.78 (rounded)
        },
        {
            name:     "multiply by zero",
            money:    money.USD.FromInt(1000),   // $10.00
            scalar:   int64(0),
            expected: money.USD.FromInt(0),      // $0.00
        },
        {
            name:     "multiply by negative",
            money:    money.USD.FromInt(1000),    // $10.00
            scalar:   int64(-2),
            expected: money.USD.FromInt(-2000),   // -$20.00
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tc.money.Mul(tc.scalar)
            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}

func TestMoneyDiv(t *testing.T) {
    cases := []struct {
        name      string
        money     money.Money
        divisor   interface{}
        mode      money.RoundingMode
        expected  money.Money
    }{
        {
            name:     "even division",
            money:    money.USD.FromInt(1000),   // $10.00
            divisor:  int64(2),
            mode:     money.RoundHalfEven,
            expected: money.USD.FromInt(500),    // $5.00
        },
        {
            name:     "division with rounding up",
            money:    money.USD.FromInt(1000),   // $10.00
            divisor:  int64(3),
            mode:     money.RoundUp,
            expected: money.USD.FromInt(334),     // $3.34 (rounded up)
        },
        {
            name:     "division with rounding half even",
            money:    money.USD.FromInt(1000),   // $10.00
            divisor:  int64(3),
            mode:     money.RoundHalfEven,
            expected: money.USD.FromInt(333),    // $3.33 (banker's)
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tc.money.Div(tc.divisor, tc.mode)
            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}

func TestMoneyCompare(t *testing.T) {
    cases := []struct {
        name     string
        a        money.Money
        b        money.Money
        expected int
    }{
        {
            name:     "equal",
            a:        money.USD.FromInt(1000),
            b:        money.USD.FromInt(1000),
            expected: 0,
        },
        {
            name:     "a less than b",
            a:        money.USD.FromInt(500),
            b:        money.USD.FromInt(1000),
            expected: -1,
        },
        {
            name:     "a greater than b",
            a:        money.USD.FromInt(1500),
            b:        money.USD.FromInt(1000),
            expected: 1,
        },
        {
            name:     "negative less than positive",
            a:        money.USD.FromInt(-1000),
            b:        money.USD.FromInt(1000),
            expected: -1,
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := tc.a.Cmp(tc.b)
            require.NoError(t, err)
            require.Equal(t, tc.expected, result)
        })
    }
}
```

### Currency Tests

```go
func TestCurrencyMismatch(t *testing.T) {
    usd := money.USD.FromInt(1000)
    eur := money.EUR.FromInt(1000)
    
    // All binary operations should fail
    operations := []struct {
        name string
        op   func() (money.Money, error)
    }{
        {"add", func() (money.Money, error) { return usd.Add(eur) }},
        {"sub", func() (money.Money, error) { return usd.Sub(eur) }},
        {"cmp", func() (money.Money, error) { 
            _, err := usd.Cmp(eur)
            return money.Money{}, err
        }},
    }
    
    for _, tc := range operations {
        t.Run(tc.name, func(t *testing.T) {
            _, err := tc.op()
            require.Error(t, err)
            
            var mismatch money.CurrencyMismatchError
            require.ErrorAs(t, err, &mismatch)
            require.Equal(t, "USD", mismatch.Left.Code())
            require.Equal(t, "EUR", mismatch.Right.Code())
        })
    }
}

func TestCurrencyParsing(t *testing.T) {
    cases := []struct {
        code     string
        expected money.Currency
        hasError bool
    }{
        {"USD", money.USD, false},
        {"EUR", money.EUR, false},
        {"usd", money.USD, false},  // Case insensitive
        {"XXX", money.Currency{}, true},  // Invalid
        {"", money.Currency{}, true},     // Empty
        {"TOOLONG", money.Currency{}, true}, // Too long
    }
    
    for _, tc := range cases {
        t.Run(tc.code, func(t *testing.T) {
            currency, err := money.ParseCurrency(tc.code)
            if tc.hasError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.expected, currency)
            }
        })
    }
}
```

---

## 2. Property-Based Tests

### Invariant Properties

```go
package money_test

import (
    "testing"
    
    "github.com/decimal/money"
    "github.com/stretchr/testify/require"
    "pgregory.net/rapid"
)

func TestProperty_AddSubIdentity(t *testing.T) {
    // For any Money m: m + 0 = m and m - 0 = m
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        zero := m.Currency().Zero()
        
        result, err := m.Add(zero)
        require.NoError(t, err)
        require.Equal(t, m, result, "m + 0 should equal m")
        
        result, err = m.Sub(zero)
        require.NoError(t, err)
        require.Equal(t, m, result, "m - 0 should equal m")
    })
}

func TestProperty_AddCommutative(t *testing.T) {
    // For any Money a, b: a + b = b + a
    rapid.Check(t, func(t *rapid.T) {
        a := genMoney(t)
        b := genMoney(t)
        
        // Skip if different currencies
        if a.Currency() != b.Currency() {
            t.Skip()
        }
        
        resultAB, err := a.Add(b)
        require.NoError(t, err)
        
        resultBA, err := b.Add(a)
        require.NoError(t, err)
        
        require.Equal(t, resultAB, resultBA, "a + b should equal b + a")
    })
}

func TestProperty_AddAssociative(t *testing.T) {
    // For any Money a, b, c: (a + b) + c = a + (b + c)
    rapid.Check(t, func(t *rapid.T) {
        a := genMoney(t)
        b := genMoney(t)
        c := genMoney(t)
        
        if a.Currency() != b.Currency() || b.Currency() != c.Currency() {
            t.Skip()
        }
        
        ab, err := a.Add(b)
        require.NoError(t, err)
        
        result1, err := ab.Add(c)
        require.NoError(t, err)
        
        bc, err := b.Add(c)
        require.NoError(t, err)
        
        result2, err := a.Add(bc)
        require.NoError(t, err)
        
        require.Equal(t, result1, result2, "(a + b) + c should equal a + (b + c)")
    })
}

func TestProperty_MulDistributesOverAdd(t *testing.T) {
    // For any Money a, b and scalar n: (a + b) * n = a*n + b*n
    rapid.Check(t, func(t *rapid.T) {
        a := genMoney(t)
        b := genMoney(t)
        n := rapid.Int64Range(-100, 100).Draw(t, "n")
        
        if n == 0 {
            t.Skip()
        }
        
        if a.Currency() != b.Currency() {
            t.Skip()
        }
        
        ab, err := a.Add(b)
        require.NoError(t, err)
        
        result1, err := ab.Mul(n)
        require.NoError(t, err)
        
        aMul, err := a.Mul(n)
        require.NoError(t, err)
        
        bMul, err := b.Mul(n)
        require.NoError(t, err)
        
        result2, err := aMul.Add(bMul)
        require.NoError(t, err)
        
        require.Equal(t, result1, result2, "(a + b) * n should equal a*n + b*n")
    })
}

func TestProperty_NegationIdentity(t *testing.T) {
    // For any Money m: m + (-m) = 0
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        neg := m.Neg()
        
        result, err := m.Add(neg)
        require.NoError(t, err)
        
        zero := m.Currency().Zero()
        require.Equal(t, zero, result, "m + (-m) should equal 0")
    })
}

func TestProperty_AbsNonNegative(t *testing.T) {
    // For any Money m: abs(m) >= 0
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        abs := m.Abs()
        
        require.True(t, abs.IsZero() || abs.IsPositive(), 
            "abs(m) should be zero or positive")
    })
}

func TestProperty_SplitPreservesTotal(t *testing.T) {
    // For any Money m split into n parts: sum(parts) = m
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        n := rapid.IntRange(1, 100).Draw(t, "n")
        
        parts, err := m.Split(n)
        require.NoError(t, err)
        require.Len(t, parts, n)
        
        // Calculate sum
        sum := m.Currency().Zero()
        for _, p := range parts {
            sum, err = sum.Add(p)
            require.NoError(t, err)
        }
        
        require.Equal(t, m, sum, "sum of split parts should equal original")
    })
}

func TestProperty_AllocationPreservesTotal(t *testing.T) {
    // For any Money m allocated by ratios: sum(parts) = m
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        ratios := genRatios(t)
        
        parts, err := m.AllocateRatios(ratios)
        require.NoError(t, err)
        
        // Calculate sum
        sum := m.Currency().Zero()
        for _, p := range parts {
            sum, err = sum.Add(p)
            require.NoError(t, err)
        }
        
        require.Equal(t, m, sum, "sum of allocated parts should equal original")
    })
}

func TestProperty_MulThenDivRoughlyIdentity(t *testing.T) {
    // For Money m and scalar n != 0: (m * n) / n ≈ m (within rounding)
    rapid.Check(t, func(t *rapid.T) {
        m := genMoney(t)
        n := rapid.Int64Range(-100, 100).Draw(t, "n")
        
        if n == 0 {
            t.Skip()
        }
        
        mulResult, err := m.Mul(n)
        require.NoError(t, err)
        
        divResult, err := mulResult.Div(n, money.RoundHalfEven)
        require.NoError(t, err)
        
        // Due to rounding, we may not get exactly m back
        // But the difference should be small
        diff, err := divResult.Sub(m)
        require.NoError(t, err)
        
        // Difference should be at most 1 cent
        oneCent := m.Currency().FromInt(1)
        require.True(t, diff.Abs().LessThan(oneCent) || diff.Abs().Equal(oneCent),
            "m * n / n should be close to m")
    })
}
```

### Generator Functions

```go
// genMoney generates random Money values for property testing
func genMoney(t *rapid.T) money.Money {
    currencies := []money.Currency{
        money.USD, money.EUR, money.GBP, money.JPY,
    }
    
    currency := currencies[rapid.IntRange(0, len(currencies)-1).Draw(t, "currency")]
    
    // Generate amount in cents (or smallest unit)
    // Range: -1,000,000,000 to 1,000,000,000 (reasonable range)
    cents := rapid.Int64Range(-100000000000, 100000000000).Draw(t, "cents")
    
    return currency.FromInt(cents)
}

// genRatios generates valid allocation ratios
func genRatios(t *rapid.T) []int {
    n := rapid.IntRange(2, 5).Draw(t, "num_parts")
    ratios := make([]int, n)
    total := 0
    
    for i := 0; i < n; i++ {
        // Each ratio 1-100, last one may be adjusted
        ratios[i] = rapid.IntRange(1, 100).Draw(t, "ratio")
        total += ratios[i]
    }
    
    return ratios
}

// genDecimal generates random Decimal values
func genDecimal(t *rapid.T) money.Decimal {
    // Generate mantissa and exponent
    mantissa := rapid.Int64Range(-1000000000, 1000000000).Draw(t, "mantissa")
    exponent := rapid.Int32Range(-10, 10).Draw(t, "exponent")
    
    return money.Decimal.NewWithMantissa(mantissa, exponent)
}
```

---

## 3. Financial Scenario Tests

### Tax Calculation Scenarios

```go
package money_test

import (
    "testing"
    "github.com/decimal/money"
)

func TestScenario_SalesTaxCalculation(t *testing.T) {
    // Common tax scenarios based on real business requirements
    
    cases := []struct {
        name           string
        items          []money.Money
        taxRate        money.Decimal  // e.g., 0.0825 for 8.25%
        expectedTotal  string
        expectedTax    string
    }{
        {
            name: "single item California",
            items: []money.Money{
                money.USD.FromString("29.99"),
            },
            taxRate:     money.Decimal.NewFloat(0.0825),
            expectedTotal: "$32.49",
            expectedTax: "$2.50",
        },
        {
            name: "multiple items Texas",
            items: []money.Money{
                money.USD.FromString("19.99"),
                money.USD.FromString("14.99"),
                money.USD.FromString("9.99"),
            },
            taxRate:      money.Decimal.NewFloat(0.0825),
            expectedTotal: "$49.26",
            expectedTax:  "$3.79",
        },
        {
            name: "Oregon (no sales tax)",
            items: []money.Money{
                money.USD.FromString("100.00"),
            },
            taxRate:      money.Decimal.NewFloat(0),
            expectedTotal: "$100.00",
            expectedTax:  "$0.00",
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            // Calculate subtotal
            var subtotal money.Money
            for _, item := range tc.items {
                if subtotal.IsZero() {
                    subtotal = item
                } else {
                    subtotal, _ = subtotal.Add(item)
                }
            }
            
            // Calculate tax (rounded per item or total)
            taxPercent := tc.taxRate.Mul(money.Decimal.NewInt(100))
            tax, err := subtotal.Mul(taxPercent)
            require.NoError(t, err)
            tax = tax.Round(money.RoundHalfUp, 2)
            
            // Calculate total
            total, err := subtotal.Add(tax)
            require.NoError(t, err)
            
            require.Equal(t, tc.expectedTotal, total.String())
            require.Equal(t, tc.expectedTax, tax.String())
        })
    }
}

func TestScenario_InterestCalculation(t *testing.T) {
    // Loan interest calculations (simplified monthly interest)
    
    cases := []struct {
        name           string
        principal      string
        annualRate     string  // e.g., "0.0599" for 5.99%
        months         int
        rounding       money.RoundingMode
        expectedPayment string
    }{
        {
            name:           "30 year mortgage",
            principal:      "300000.00",
            annualRate:     "0.0599",
            months:         360,
            rounding:       money.RoundHalfEven,
            expectedPayment: "$1,793.24",
        },
        {
            name:           "5 year auto loan",
            principal:      "25000.00",
            annualRate:     "0.0499",
            months:         60,
            rounding:       money.RoundHalfEven,
            expectedPayment: "$466.38",
        },
        {
            name:           "simple interest",
            principal:      "1000.00",
            annualRate:     "0.10",
            months:         12,
            rounding:       money.RoundHalfUp,
            expectedPayment: "$91.67",  // $1100 / 12
        },
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            principal := money.USD.FromString(tc.principal)
            rate := money.Decimal.NewString(tc.annualRate)
            
            // Monthly payment formula: P * [r(1+r)^n] / [(1+r)^n - 1]
            monthlyRate := rate.Div(money.Decimal.NewInt(12))
            
            // (1 + r)^n
            one := money.Decimal.NewInt(1)
            factor := monthlyRate.Add(one)
            compound := factor.Exp(int64(tc.months))  // Simplified
            
            // numerator: r * (1+r)^n
            numerator := monthlyRate.Mul(compound)
            
            // denominator: (1+r)^n - 1
            denominator := compound.Sub(one)
            
            // ratio: numerator / denominator
            ratio := numerator.Div(denominator)
            
            // payment: principal * ratio
            payment := principal.Mul(ratio)
            payment = payment.Round(tc.rounding, 2)
            
            require.Equal(t, tc.expectedPayment, payment.String())
        })
    }
}

func TestScenario_MultiCurrencySettlement(t *testing.T) {
    // Test currency conversion in settlement scenarios
    
    t.Run("USD to EUR conversion", func(t *testing.T) {
        usdAmount := money.USD.FromString("1000.00")
        rate := money.Decimal.NewFloat(0.92)  // 1 USD = 0.92 EUR
        
        eurAmount, err := usdAmount.Convert(money.EUR, rate, money.RoundHalfEven)
        require.NoError(t, err)
        
        require.Equal(t, "EUR", eurAmount.Currency().Code())
        require.Equal(t, "€920.00", eurAmount.String())
    })
    
    t.Run("conversion with inverse rate", func(t *testing.T) {
        usdAmount := money.USD.FromString("1000.00")
        rate := money.Decimal.NewFloat(1.08)  // EUR/USD rate
        
        eurAmount, err := usdAmount.Convert(money.EUR, rate, money.RoundHalfEven)
        require.NoError(t, err)
        
        // Convert back
        backToUSD, err := eurAmount.Convert(money.USD, money.Decimal.NewFloat(1.0).Div(rate), money.RoundHalfEven)
        require.NoError(t, err)
        
        // Should be close to original (within rounding)
        diff, _ := usdAmount.Sub(backToUSD)
        require.True(t, diff.Abs().LessThan(money.USD.FromInt(2)))  // Within 2 cents
    })
}

func TestScenario_InvoiceAllocation(t *testing.T) {
    // Test fair allocation of invoice across line items
    
    t.Run("allocate discount proportionally", func(t *testing.T) {
        // Invoice with items totaling $100, get $10 discount
        invoice := money.USD.FromString("100.00")
        discount := money.USD.FromString("10.00")
        
        // Items before discount
        items := []money.Money{
            money.USD.FromString("50.00"),  // 50%
            money.USD.FromString("30.00"),  // 30%
            money.USD.FromString("20.00"),  // 20%
        }
        
        // Allocate discount proportionally
        discountRatios := []int{50, 30, 20}
        discountedItems, err := invoice.AllocateDiscount(discount, discountRatios, money.RoundHalfEven)
        require.NoError(t, err)
        
        require.Len(t, discountedItems, 3)
        require.Equal(t, "$45.00", discountedItems[0].String())  // $50 - $5
        require.Equal(t, "$27.00", discountedItems[1].String())  // $30 - $3
        require.Equal(t, "$18.00", discountedItems[2].String())  // $20 - $2
    })
}
```

### Edge Case Financial Scenarios

```go
func TestScenario_ZeroCurrencyHandling(t *testing.T) {
    t.Run("zero amounts in calculations", func(t *testing.T) {
        zero := money.USD.Zero()
        normal := money.USD.FromString("100.00")
        
        // Zero + normal = normal
        result, err := zero.Add(normal)
        require.NoError(t, err)
        require.Equal(t, normal, result)
        
        // Normal - normal = zero
        result, err = normal.Sub(normal)
        require.NoError(t, err)
        require.Equal(t, zero, result)
        
        // Zero * anything = zero
        result, err = zero.Mul(1000000)
        require.NoError(t, err)
        require.Equal(t, zero, result)
    })
    
    t.Run("very small amounts", func(t *testing.T) {
        // Test handling of sub-cent amounts
        tiny := money.USD.FromInt(1)  // $0.01
        halfTiny := money.USD.FromFloat(0.005)
        
        // Division that results in tiny amounts
        result, err := tiny.Div(3, money.RoundHalfEven)
        require.NoError(t, err)
        require.Equal(t, "$0.00", result.String())
    })
}

func TestScenario_NegativeMoneyHandling(t *testing.T) {
    t.Run("refunds and chargebacks", func(t *testing.T) {
        original := money.USD.FromString("100.00")
        refund := money.USD.FromString("-100.00")
        
        total, err := original.Add(refund)
        require.NoError(t, err)
        require.True(t, total.IsZero())
    })
    
    t.Run("net balance calculation", func(t *testing.T) {
        // Multiple credits and debits
        transactions := []money.Money{
            money.USD.FromString("500.00"),   // Credit
            money.USD.FromString("-100.00"),  // Debit
            money.USD.FromString("250.00"),   // Credit
            money.USD.FromString("-50.00"),   // Debit
        }
        
        var balance money.Money
        for _, tx := range transactions {
            if balance.IsZero() {
                balance = tx
            } else {
                balance, _ = balance.Add(tx)
            }
        }
        
        require.Equal(t, "$600.00", balance.String())
    })
}
```

---

## 4. Roundtrip Tests

### Serialization Roundtrips

```go
func TestRoundtrip_JSON(t *testing.T) {
    original := money.USD.FromString("12345.67")
    
    // Marshal
    data, err := json.Marshal(original)
    require.NoError(t, err)
    
    // Unmarshal
    var restored money.Money
    err = json.Unmarshal(data, &restored)
    require.NoError(t, err)
    
    require.Equal(t, original, restored)
}

func TestRoundtrip_String(t *testing.T) {
    original := money.USD.FromString("12345.67")
    
    // Parse what String() produces
    str := original.String()
    restored, err := money.USD.FromString(str)
    require.NoError(t, err)
    
    require.Equal(t, original, restored)
}
```

---

## 5. Wasm Integration Tests

### JavaScript Tests

```javascript
// tests/money.spec.js

import { describe, it, expect, beforeAll } from 'vitest';
import init, { Money, Decimal, RoundingMode, Error } from '../dist/decimal_wasm.js';

beforeAll(async () => {
    await init();
});

describe('Money', () => {
    describe('construction', () => {
        it('should create from string', () => {
            const money = new Money('USD', '19.99');
            expect(money.toString()).toBe('$19.99');
        });
        
        it('should create from int', () => {
            const money = Money.fromInt('USD', 1999);
            expect(money.toString()).toBe('$19.99');
        });
    });
    
    describe('arithmetic', () => {
        it('should add correctly', () => {
            const a = new Money('USD', '10.00');
            const b = new Money('USD', '5.00');
            const result = a.add(b);
            expect(result.toString()).toBe('$15.00');
        });
        
        it('should throw on currency mismatch', () => {
            const a = new Money('USD', '10.00');
            const b = new Money('EUR', '5.00');
            expect(() => a.add(b)).toThrow();
        });
        
        it('should divide with rounding', () => {
            const money = new Money('USD', '10.00');
            const result = money.divInt(3, RoundingMode.HALF_EVEN);
            expect(result.toString()).toBe('$3.33');
        });
    });
    
    describe('allocation', () => {
        it('should split evenly', () => {
            const money = new Money('USD', '10.00');
            const parts = money.split(3);
            expect(parts.length).toBe(3);
            // 10 / 3 = 3.333..., so: [3.34, 3.33, 3.33]
            expect(parts[0].toString()).toBe('$3.34');
            expect(parts[1].toString()).toBe('$3.33');
            expect(parts[2].toString()).toBe('$3.33');
        });
    });
});

describe('Property Tests (JS)', () => {
    it('should satisfy add-sub identity', () => {
        const money = new Money('USD', '123.45');
        const zero = Money.fromInt('USD', 0);
        
        const result = money.add(zero);
        expect(result.toString()).toBe(money.toString());
    });
    
    it('should preserve total on split', () => {
        const original = new Money('USD', '100.00');
        const parts = original.split(3);
        
        let sum = Money.fromInt('USD', 0);
        for (const part of parts) {
            sum = sum.add(part);
        }
        
        expect(sum.toString()).toBe(original.toString());
    });
});
```

---

## 6. Test Coverage Targets

| Component | Coverage Target |
|-----------|-----------------|
| Core arithmetic (add, sub, mul, div) | 100% |
| Rounding modes | 100% |
| Currency handling | 100% |
| Error types | 100% |
| Allocation/split | 100% |
| String parsing | 95% |
| Serialization | 95% |
| Conversion | 90% |

---

## Summary

| Test Type | Purpose | Tools |
|-----------|---------|-------|
| Unit Tests | Verify individual operations | Go test, Vitest |
| Property Tests | Verify mathematical invariants | rapid, fast-check |
| Scenario Tests | Verify real-world financial use cases | Custom test suite |
| Integration Tests | Verify Wasm/JS interop | Vitest, wasm-bindgen-test |
| Roundtrip Tests | Verify serialization consistency | JSON, String tests |