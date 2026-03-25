# Benchmark Plan

## Overview

This document outlines a comprehensive benchmarking strategy to measure the performance of the decimal money library against established competitors, identify optimization opportunities, and establish performance baselines.

---

## Benchmark Objectives

1. **Competitive Analysis**: Compare performance against `shopspring/decimal` (Go) and `decimal.js`/`big.js` (JavaScript)
2. **Hot Path Identification**: Find the most performance-critical operations
3. **Optimization Tracking**: Measure impact of code changes over time
4. **Resource Profiling**: Measure memory allocation and CPU time
5. **Edge Case Performance**: Measure performance at boundary conditions

---

## Benchmark Environment

### Hardware Specifications

```
CPU: Intel Core i9-12900K (or equivalent)
Memory: 64GB DDR5
Storage: NVMe SSD
OS: Linux (Ubuntu 22.04 LTS)
```

### Software Versions

```
Go: 1.21+
Rust: 1.70+
Node.js: 20 LTS
```

### Measurement Methodology

1. **Warm-up**: 1000 iterations before measurement
2. **Iterations**: Minimum 10,000 iterations per benchmark
3. **GC**: Disable GC during benchmarks (Go: `debug.SetGCPercent(-1)`)
4. **Stabilization**: Run until variance < 5%
5. **Statistical Significance**: Use p95 confidence interval

---

## Operations to Benchmark

### Core Arithmetic Operations

| Operation | Description | Priority |
|-----------|-------------|----------|
| Add | Add two Money values | Critical |
| Sub | Subtract two Money values | Critical |
| Mul | Multiply Money by scalar | Critical |
| Div | Divide Money by scalar | Critical |
| Cmp | Compare two Money values | High |
| Abs | Absolute value | Medium |
| Neg | Negation | Medium |
| Zero | Create zero value | Low |

### String Operations

| Operation | Description | Priority |
|-----------|-------------|----------|
| Parse | Parse string to Money | Critical |
| Format | Format Money to string | Critical |
| MarshalJSON | JSON serialization | Medium |
| UnmarshalJSON | JSON deserialization | Medium |

### Complex Operations

| Operation | Description | Priority |
|-----------|-------------|----------|
| Split | Split Money into n parts | High |
| Allocate | Allocate by ratios | High |
| Convert | Currency conversion | Medium |
| Round | Apply rounding | Critical |

---

## Go Benchmarks (shopspring/decimal comparison)

### Benchmark Structure

```go
package decimal_test

import (
    "testing"
    "github.com/shopspring/decimal"
    decimalmoney "github.com/decimal/money"
)

func BenchmarkAddDecimal(b *testing.B) {
    // shopspring/decimal
    a := decimal.NewFromFloat(123.45)
    b.Run("shopspring", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = a.Add(decimal.NewFromFloat(67.89))
        }
    })
    
    // Our library
    money := decimalmoney.USD.FromFloat(123.45)
    b.Run("decimal-money", func(b *testing.B) {
        other := decimalmoney.USD.FromFloat(67.89)
        for i := 0; i < b.N; i++ {
            _ = money.Add(other)
        }
    })
}
```

### Go Benchmark Suite

```go
// benchmarks/decimal_test.go

package benchmarks

import (
    "testing"
    "math/rand"
    "encoding/json"
    "strconv"
    
    decimalmoney "github.com/decimal/money"
    "github.com/shopspring/decimal"
)

var (
    testValuesUSD = generateTestValues(decimalmoney.USD, 1000)
    testValuesEUR = generateTestValues(decimalmoney.EUR, 1000)
)

func generateTestValues(currency decimalmoney.Currency, n int) []decimalmoney.Money {
    values := make([]decimalmoney.Money, n)
    for i := 0; i < n; i++ {
        // Generate random cent values
        cents := rand.Int63n(1000000) // up to $10,000.00
        values[i] = currency.FromInt(cents)
    }
    return values
}

// === Addition Benchmarks ===

func BenchmarkAddSameCurrency(b *testing.B) {
    a := testValuesUSD[0]
    b := testValuesUSD[1]
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result, err := a.Add(b)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

func BenchmarkAddDifferentCurrency(b *testing.B) {
    a := testValuesUSD[0]
    b := testValuesEUR[0] // Different currency
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := a.Add(b)
        if err == nil {
            b.Fatal("expected currency mismatch error")
        }
        _ = err
    }
}

// === Multiplication Benchmarks ===

func BenchmarkMulInt(b *testing.B) {
    money := decimalmoney.USD.FromFloat(123.45)
    factor := int64(3)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result, err := money.MulInt(factor)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

func BenchmarkMulDecimal(b *testing.B) {
    money := decimalmoney.USD.FromFloat(123.45)
    factor := decimalmoney.Decimal.NewFloat(3.5)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result, err := money.Mul(factor)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

// === Division Benchmarks ===

func BenchmarkDivInt(b *testing.B) {
    money := decimalmoney.USD.FromFloat(100.00)
    divisor := int64(3)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        result, err := money.DivInt(divisor, decimalmoney.RoundHalfEven)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

func BenchmarkDivWithRounding(b *testing.B) {
    money := decimalmoney.USD.FromFloat(100.00)
    divisor := int64(3)
    
    // Test different rounding modes
    modes := []decimalmoney.RoundingMode{
        decimalmoney.RoundHalfEven,
        decimalmoney.RoundUp,
        decimalmoney.RoundDown,
    }
    
    for _, mode := range modes {
        b.Run(mode.String(), func(b *testing.B) {
            b.ReportAllocs()
            for i := 0; i < b.N; i++ {
                result, err := money.DivInt(divisor, mode)
                if err != nil {
                    b.Fatal(err)
                }
                _ = result
            }
        })
    }
}

// === String Parsing Benchmarks ===

func BenchmarkParseString(b *testing.B) {
    inputs := []string{
        "0.01",
        "1.00",
        "123.45",
        "1234.56",
        "12345.67",
        "123456.78",
        "1234567.89",
    }
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        input := inputs[i%len(inputs)]
        result, err := decimalmoney.USD.FromString(input)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

func BenchmarkFormatString(b *testing.B) {
    money := decimalmoney.USD.FromFloat(12345.67)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = money.String()
    }
}

// === JSON Serialization ===

func BenchmarkMarshalJSON(b *testing.B) {
    money := decimalmoney.USD.FromFloat(12345.67)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        data, err := json.Marshal(money)
        if err != nil {
            b.Fatal(err)
        }
        _ = data
    }
}

func BenchmarkUnmarshalJSON(b *testing.B) {
    data := []byte(`"12345.67"`)
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        var money decimalmoney.Money
        err := json.Unmarshal(data, &money)
        if err != nil {
            b.Fatal(err)
        }
        _ = money
    }
}

// === Split/Allocation ===

func BenchmarkSplit(b *testing.B) {
    money := decimalmoney.USD.FromFloat(100.00)
    
    b.ReportAllAlloc()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        parts, err := money.Split(3)
        if err != nil {
            b.Fatal(err)
        }
        _ = parts
    }
}

func BenchmarkAllocateRatios(b *testing.B) {
    money := decimalmoney.USD.FromFloat(100.00)
    ratios := []int{50, 30, 20}
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        parts, err := money.AllocateRatios(ratios)
        if err != nil {
            b.Fatal(err)
        }
        _ = parts
    }
}

// === Comparison ===

func BenchmarkCompare(b *testing.B) {
    a := testValuesUSD[0]
    b := testValuesUSD[1]
    
    b.ReportAllocs()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        cmp, err := a.Cmp(b)
        if err != nil {
            b.Fatal(err)
        }
        _ = cmp
    }
}
```

---

## JavaScript Benchmarks

### Benchmark Structure

```javascript
// benchmarks/decimal.bench.js

import { bench, run } from 'tinybench';
import Decimal from 'decimal.js';
import { Money } from '../dist/decimal_wasm.js';

const suite = new bench({ time: 100, iterations: 10000 });

suite.add('decimal.js - add', () => {
    const a = new Decimal('123.45');
    const b = new Decimal('67.89');
    return a.plus(b);
});

suite.add('wasm-money - add', async () => {
    await init(); // Ensure Wasm is loaded
    const a = new Money('USD', '123.45');
    const b = new Money('USD', '67.89');
    return a.add(b);
});

// Run and report
await suite.run();
console.table(suite.table());
```

### JavaScript Benchmark Suite

```javascript
// benchmarks/index.mjs

import { bench, group, before, after } from 'mitata';
import Decimal from 'decimal.js';

// Wasm initialization
let Money, RoundingMode;
before(async () => {
    const wasm = await import('../dist/decimal_wasm.js');
    await wasm.default();
    Money = wasm.Money;
    RoundingMode = wasm.RoundingMode;
});

// === Arithmetic Operations ===

group('addition', () => {
    bench('decimal.js', () => {
        const a = new Decimal('123.45');
        const b = new Decimal('67.89');
        return a.plus(b);
    });
    
    bench('wasm-money', () => {
        const a = new Money('USD', '123.45');
        const b = new Money('USD', '67.89');
        return a.add(b);
    });
});

group('multiplication', () => {
    bench('decimal.js', () => {
        const a = new Decimal('123.45');
        return a.times(3);
    });
    
    bench('wasm-money', () => {
        const a = new Money('USD', '123.45');
        return a.mulInt(3);
    });
});

group('division with rounding', () => {
    const modes = {
        'HALF_EVEN': 0,
        'UP': 1,
        'DOWN': 2,
    };
    
    for (const [name, mode] of Object.entries(modes)) {
        bench(`decimal.js - ${name}`, () => {
            const a = new Decimal('100');
            return a.dividedBy(3).toDecimalPlaces(2, mode);
        });
        
        bench(`wasm-money - ${name}`, () => {
            const a = new Money('USD', '100');
            return a.divInt(3, mode);
        });
    }
});

// === String Operations ===

group('parsing', () => {
    const testStrings = [
        '0.01', '1.00', '123.45', '1234.56', 
        '12345.67', '123456.78', '1234567.89'
    ];
    
    bench('decimal.js', () => {
        for (const s of testStrings) {
            new Decimal(s);
        }
    });
    
    bench('wasm-money', () => {
        for (const s of testStrings) {
            new Money('USD', s);
        }
    });
});

group('formatting', () => {
    bench('decimal.js', () => {
        const d = new Decimal('12345.67');
        return d.toFixed(2);
    });
    
    bench('wasm-money', () => {
        const m = new Money('USD', '12345.67');
        return m.toString();
    });
});

// === Allocation ===

group('split (100 / 3)', () => {
    bench('decimal.js', () => {
        const total = new Decimal('100');
        const n = 3;
        const base = total.dividedBy(n).toDecimalPlaces(2, Decimal.ROUND_DOWN);
        const remainder = total.minus(base.times(n));
        
        const parts = [];
        for (let i = 0; i < n; i++) {
            parts.push(i < remainder ? base.plus(0.01) : base);
        }
        return parts;
    });
    
    bench('wasm-money', () => {
        const total = new Money('USD', '100');
        return total.split(3);
    });
});

await run();
```

---

## Metrics to Collect

### Primary Metrics

| Metric | Description | Units |
|--------|-------------|-------|
| Throughput | Operations per second | ops/sec |
| Latency Mean | Average operation time | ns/op |
| Latency p50 | Median latency | ns/op |
| Latency p95 | 95th percentile | ns/op |
| Latency p99 | 99th percentile | ns/op |
| Allocations | Objects allocated per op | objects/op |
| Allocation Size | Bytes allocated per op | bytes/op |

### Secondary Metrics

| Metric | Description |
|--------|-------------|
| Memory Working Set | Memory used during benchmark |
| CPU Time | Total CPU time consumed |
| GC Pauses | Garbage collection pauses (Go) |
| Wasm Module Size | Size of compiled Wasm binary |
| Parse Time | Time to parse string input |
| Format Time | Time to format output string |

---

## Edge Case Benchmarks

### Catastrophic Cancellation

```go
func BenchmarkCatastrophicCancellation(b *testing.B) {
    // Large values with small differences
    a := USD.FromFloat(1000000.00)
    b := USD.FromFloat(999999.99)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        diff, _ := a.Sub(b)
        _ = diff
    }
}
```

### Precision Boundaries

```go
func BenchmarkPrecisionBoundary(b *testing.B) {
    // Exactly at decimal boundary
    tests := []struct {
        value    string
        expected string
    }{
        {"0.005", "0.01"},  // Round to 2 decimals
        {"0.015", "0.02"},  // Round to 2 decimals (tie)
        {"0.995", "1.00"},  // Carry over
        {"999999.995", "1000000.00"},  // Large with carry
    }
    
    for _, tc := range tests {
        money, _ := USD.FromString(tc.value)
        b.Run(tc.value, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                rounded := money.Round(RoundHalfEven, 2)
                _ = rounded
            }
        })
    }
}
```

### Maximum Values

```go
func BenchmarkMaxValueArithmetic(b *testing.B) {
    max := USD.FromInt(math.MaxInt64)
    factor := USD.FromInt(2)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := max.Mul(factor)
        _ = err // Should overflow
    }
}

func BenchmarkNearMaxValue(b *testing.B) {
    // Just under overflow threshold
    nearMax := USD.FromInt(math.MaxInt64 - 1)
    factor := USD.FromInt(2)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result, err := nearMax.Mul(factor)
        _ = result
        _ = err
    }
}
```

### Division Edge Cases

```go
func BenchmarkDivisionEdgeCases(b *testing.B) {
    tests := []struct {
        name   string
        dividend string
        divisor  int64
    }{
        {"even_divide", "100.00", 4},
        {"repeating", "100.00", 3},
        {"very_small_result", "1.00", 1000},
        {"near_zero_dividend", "0.01", 3},
    }
    
    for _, tc := range tests {
        money, _ := USD.FromString(tc.dividend)
        b.Run(tc.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                result, err := money.DivInt(tc.divisor, RoundHalfEven)
                _ = result
                _ = err
            }
        })
    }
}
```

---

## Reporting Format

### JSON Report Structure

```json
{
  "metadata": {
    "timestamp": "2024-01-15T10:30:00Z",
    "commit": "abc123",
    "branch": "main",
    "runner": {
      "cpu": "Intel Core i9-12900K",
      "memory": "64GB DDR5",
      "os": "Ubuntu 22.04"
    }
  },
  "results": {
    "addition": {
      "shopspring_decimal": {
        "ops_per_sec": 12500000,
        "mean_ns_op": 80,
        "p50_ns": 78,
        "p95_ns": 95,
        "p99_ns": 110,
        "allocs_per_op": 2,
        "bytes_per_op": 48
      },
      "decimal_money": {
        "ops_per_sec": 45000000,
        "mean_ns_op": 22,
        "p50_ns": 21,
        "p95_ns": 28,
        "p99_ns": 35,
        "allocs_per_op": 0,
        "bytes_per_op": 0
      },
      "speedup": 3.6
    }
  }
}
```

### Comparison Chart (ASCII)

```
Operation        shopspring    our-lib    speedup
──────────────────────────────────────────────────
Add                 80 ns      22 ns       3.6x
Sub                 82 ns      21 ns       3.9x
Mul                120 ns      45 ns       2.7x
Div                150 ns      68 ns       2.2x
Parse (string)     450 ns     120 ns       3.8x
Format (string)    200 ns      55 ns       3.6x
```

---

## Continuous Benchmarking

### GitHub Actions Integration

```yaml
# .github/workflows/benchmarks.yml
name: Benchmarks

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Run Go benchmarks
        run: |
          go test -bench=. -benchmem -count=5 \
            ./benchmarks/... | tee benchmark.txt
            
      - name: Set up Node
        uses: actions/setup-node@v3
        with:
          node-version: 20
          
      - name: Run JS benchmarks
        run: node benchmarks/index.mjs > js_benchmark.json
        
      - name: Generate comparison report
        run: |
          go run scripts/compare_benchmarks.go \
            --go benchmark.txt \
            --js js_benchmark.json \
            --output report.html
            
      - name: Upload report
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-report
          path: report.html
```

---

## Performance Regression Detection

### Thresholds

```yaml
# .benchmarks/thresholds.yaml
regression_thresholds:
  # If any operation slows by more than 10%, fail CI
  max_latency_increase_percent: 10
  
  # If any operation increases allocations, fail CI  
  max_allocation_increase_percent: 5
  
  # If memory usage increases significantly, warn
  max_memory_increase_percent: 20

comparison:
  baseline: main  # Compare against main branch
  alert_on_regression: true
```

---

## Summary

| Category | Operations | Focus |
|----------|------------|-------|
| Core Arithmetic | Add, Sub, Mul, Div | Throughput, allocations |
| Comparison | Cmp, Eq, Lt, Gt | Latency |
| String | Parse, Format | Speed, allocation |
| Complex | Split, Allocate | Algorithm efficiency |
| Edge Cases | Overflow, boundaries | Correctness under stress |
| Serialization | JSON, etc. | Size, speed |