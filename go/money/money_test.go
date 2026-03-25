package money

import (
	"math"
	"testing"
)

func TestMoneyFromInt(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		amount   int64
		expected int64
	}{
		{"USD 1999 cents", USD, 1999, 1999},
		{"USD zero", USD, 0, 0},
		{"USD negative", USD, -500, -500},
		{"JPY 0 exponent", JPY, 1000, 1000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.currency.FromInt(tc.amount)
			if m.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, m.Amount())
			}
			if m.Currency() != tc.currency {
				t.Errorf("expected currency %v, got %v", tc.currency, m.Currency())
			}
		})
	}
}

func TestMoneyString(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency Currency
		expected string
	}{
		{"USD positive", 1999, USD, "USD 19.99"},
		{"USD zero", 0, USD, "USD 0.00"},
		{"USD negative", -1999, USD, "USD -19.99"},
		{"JPY", 1999, JPY, "JPY 1999"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.currency.FromInt(tc.amount)
			if m.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, m.String())
			}
		})
	}
}

func TestMoneyAdd(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int64
	}{
		{"simple", 1000, 500, 1500},
		{"with carry", 999, 2, 1001},
		{"negative", -500, 1000, 500},
		{"both negative", -1000, -500, -1500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := USD.FromInt(tc.a)
			b := USD.FromInt(tc.b)
			result, err := a.Add(b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Amount())
			}
		})
	}
}

func TestMoneyAddCurrencyMismatch(t *testing.T) {
	usd := USD.FromInt(1000)
	eur := EUR.FromInt(1000)

	_, err := usd.Add(eur)
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}

	mismatch, ok := err.(CurrencyMismatchError)
	if !ok {
		t.Fatalf("expected CurrencyMismatchError, got %T", err)
	}
	if mismatch.Left != USD || mismatch.Right != EUR {
		t.Errorf("unexpected currencies: %v and %v", mismatch.Left, mismatch.Right)
	}
}

func TestMoneySub(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int64
	}{
		{"simple", 1000, 300, 700},
		{"negative result", 100, 500, -400},
		{"from zero", 0, 500, -500},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := USD.FromInt(tc.a)
			b := USD.FromInt(tc.b)
			result, err := a.Sub(b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Amount())
			}
		})
	}
}

func TestMoneyMulInt(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		factor   int64
		expected int64
	}{
		{"simple", 1000, 3, 3000},
		{"by zero", 1000, 0, 0},
		{"by negative", 1000, -2, -2000},
		{"by one", 1000, 1, 1000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := USD.FromInt(tc.amount)
			result, err := m.MulInt(tc.factor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Amount())
			}
		})
	}
}

func TestMoneyMulIntOverflow(t *testing.T) {
	m := USD.FromInt(math.MaxInt64)
	_, err := m.MulInt(2)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestMoneyDivInt(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		divisor  int64
		mode     RoundingMode
		expected int64
	}{
		{"even division", 1000, 2, RoundHalfEven, 500},
		{"round up", 1000, 3, RoundUp, 334},
		{"round half even 1.333", 1000, 3, RoundHalfEven, 333},
		{"round half even 1.335", 1005, 3, RoundHalfEven, 335},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := USD.FromInt(tc.amount)
			result, err := m.DivInt(tc.divisor, tc.mode)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Amount())
			}
		})
	}
}

func TestMoneyDivIntByZero(t *testing.T) {
	m := USD.FromInt(1000)
	_, err := m.DivInt(0, RoundHalfEven)
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

func TestMoneyCmp(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int
	}{
		{"equal", 1000, 1000, 0},
		{"a less", 500, 1000, -1},
		{"a greater", 1500, 1000, 1},
		{"negative", -1000, 0, -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := USD.FromInt(tc.a)
			b := USD.FromInt(tc.b)
			result, err := a.Cmp(b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestMoneySplit(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		n        int
		expected []int64
	}{
		{"split 3 ways 100", 10000, 3, []int64{3334, 3333, 3333}},
		{"split 2 ways", 1000, 2, []int64{500, 500}},
		{"split 1 way", 1000, 1, []int64{1000}},
		{"split with remainder", 1001, 3, []int64{334, 334, 333}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := USD.FromInt(tc.amount)
			result, err := m.Split(tc.n)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tc.n {
				t.Fatalf("expected %d parts, got %d", tc.n, len(result))
			}

			// Check sum equals original
			var sum int64
			for i, p := range result {
				sum += p.Amount()
				if p.Amount() != tc.expected[i] {
					t.Errorf("part %d: expected %d, got %d", i, tc.expected[i], p.Amount())
				}
			}
			if sum != tc.amount {
				t.Errorf("sum %d != original %d", sum, tc.amount)
			}
		})
	}
}

func TestMoneyAllocateRatios(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		ratios   []int
		expected []int64
	}{
		{"5:3:2", 10000, []int{5, 3, 2}, []int64{5000, 3000, 2000}},
		{"equal split", 9000, []int{1, 1, 1}, []int64{3000, 3000, 3000}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := USD.FromInt(tc.amount)
			result, err := m.AllocateRatios(tc.ratios)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tc.ratios) {
				t.Fatalf("expected %d parts, got %d", len(tc.ratios), len(result))
			}

			// Check sum equals original
			var sum int64
			for i, p := range result {
				sum += p.Amount()
				if p.Amount() != tc.expected[i] {
					t.Errorf("part %d: expected %d, got %d", i, tc.expected[i], p.Amount())
				}
			}
			if sum != tc.amount {
				t.Errorf("sum %d != original %d", sum, tc.amount)
			}
		})
	}
}

func TestMoneyFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		currency Currency
		expected int64
		hasError bool
	}{
		{"simple", "19.99", USD, 1999, false},
		{"zero", "0.00", USD, 0, false},
		{"negative", "-5.00", USD, -500, false},
		{"integer only", "10", USD, 1000, false},
		{"too many decimals", "10.123", USD, 1012, false}, // truncated to 10.12
		{"invalid chars", "abc", USD, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.currency.FromString(tc.input)
			if tc.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Amount() != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result.Amount())
			}
		})
	}
}

func TestMoneyIsZero(t *testing.T) {
	m := USD.FromInt(0)
	if !m.IsZero() {
		t.Error("expected IsZero() to be true")
	}

	m = USD.FromInt(1)
	if m.IsZero() {
		t.Error("expected IsZero() to be false")
	}
}

func TestMoneyIsPositive(t *testing.T) {
	m := USD.FromInt(1)
	if !m.IsPositive() {
		t.Error("expected IsPositive() to be true")
	}

	m = USD.FromInt(0)
	if m.IsPositive() {
		t.Error("expected IsPositive() to be false for zero")
	}

	m = USD.FromInt(-1)
	if m.IsPositive() {
		t.Error("expected IsPositive() to be false for negative")
	}
}

func TestMoneyIsNegative(t *testing.T) {
	m := USD.FromInt(-1)
	if !m.IsNegative() {
		t.Error("expected IsNegative() to be true")
	}

	m = USD.FromInt(0)
	if m.IsNegative() {
		t.Error("expected IsNegative() to be false for zero")
	}

	m = USD.FromInt(1)
	if m.IsNegative() {
		t.Error("expected IsNegative() to be false for positive")
	}
}

func TestMoneyNeg(t *testing.T) {
	m := USD.FromInt(1000)
	neg := m.Neg()
	if neg.Amount() != -1000 {
		t.Errorf("expected -1000, got %d", neg.Amount())
	}

	// Double negation
	neg2 := neg.Neg()
	if neg2.Amount() != 1000 {
		t.Errorf("expected 1000, got %d", neg2.Amount())
	}
}

func TestMoneyAbs(t *testing.T) {
	m := USD.FromInt(-1000)
	abs := m.Abs()
	if abs.Amount() != 1000 {
		t.Errorf("expected 1000, got %d", abs.Amount())
	}

	m = USD.FromInt(1000)
	abs = m.Abs()
	if abs.Amount() != 1000 {
		t.Errorf("expected 1000, got %d", abs.Amount())
	}
}

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		code     string
		expected Currency
		hasError bool
	}{
		{"USD", USD, false},
		{"EUR", EUR, false},
		{"usd", USD, false}, // case insensitive
		{"XXX", Currency{}, true},
		{"", Currency{}, true},
		{"TOOLONG", Currency{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			currency, err := ParseCurrency(tc.code)
			if tc.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if currency != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, currency)
			}
		})
	}
}

func TestRoundingModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     RoundingMode
		input    int64
		scale    int
		target   int
		expected int64
	}{
		// 1.001 rounded to 2 decimals: quotient=100, remainder=1, RoundUp increments to 101 (which is 1.01)
		{"RoundUp 1.001", RoundUp, 1001, 3, 2, 101},
		// 1.999 rounded to 2 decimals: quotient=199, remainder=9, RoundDown keeps 199 (which is 1.99)
		{"RoundDown 1.999", RoundDown, 1999, 3, 2, 199},
		// 1.015 rounded to 2 decimals: quotient=101, remainder=5, RoundHalfUp increments to 102 (1.02)
		{"RoundHalfUp 1.015", RoundHalfUp, 1015, 3, 2, 102},
		// 1.015 rounded to 2 decimals: quotient=101, remainder=5, RoundHalfDown stays 101 (1.01)
		{"RoundHalfDown 1.015", RoundHalfDown, 1015, 3, 2, 101},
		// 1.015 rounded to 2 decimals: quotient=101, remainder=5, quotient is odd, RoundHalfEven increments to 102 (1.02)
		{"RoundHalfEven 1.015", RoundHalfEven, 1015, 3, 2, 102},
		// 1.025 rounded to 2 decimals: quotient=102, remainder=5, quotient is even, RoundHalfEven keeps 102 (1.02)
		{"RoundHalfEven 1.025", RoundHalfEven, 1025, 3, 2, 102},
		// 1.001 rounded to 2 decimals (ceiling): positive with remainder, ceiling rounds up to 101 (1.01)
		{"RoundCeiling 1.001", RoundCeiling, 1001, 3, 2, 101},
		// 1.001 rounded to 2 decimals (floor): positive with remainder, floor keeps 100 (1.00)
		{"RoundFloor 1.001", RoundFloor, 1001, 3, 2, 100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.mode.round(tc.input, tc.scale, tc.target)
			if result != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestDecimalFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{"simple", "19.99", "19.99", false},
		{"negative", "-5.5", "-5.5", false},
		{"integer", "100", "100", false},
		{"leading zero", "0.5", "0.5", false},
		{"empty", "", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewDecimalFromString(tc.input)
			if tc.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, d.String())
			}
		})
	}
}

func TestDecimalAdd(t *testing.T) {
	a, _ := NewDecimalFromString("1.5")
	b, _ := NewDecimalFromString("2.3")
	result := a.Add(b)
	if result.String() != "3.8" {
		t.Errorf("expected 3.8, got %s", result.String())
	}
}

func TestDecimalSub(t *testing.T) {
	a, _ := NewDecimalFromString("5.0")
	b, _ := NewDecimalFromString("3.0")
	result := a.Sub(b)
	if result.String() != "2" && result.String() != "2.0" {
		t.Errorf("expected 2 or 2.0, got %s", result.String())
	}
}

func TestDecimalMul(t *testing.T) {
	a, _ := NewDecimalFromString("2.5")
	b, _ := NewDecimalFromString("3.0")
	result := a.Mul(b)
	if result.String() != "7.50" {
		t.Errorf("expected 7.50, got %s", result.String())
	}
}

func TestDecimalDiv(t *testing.T) {
	a, _ := NewDecimalFromString("10.0")
	b, _ := NewDecimalFromString("3.0")
	result, err := a.Div(b, RoundHalfEven)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10.0 / 3.0 ≈ 3.33, scale should be 1
	if result.Scale() != 1 {
		t.Errorf("expected scale 1, got %d", result.Scale())
	}
}

func TestDecimalCmp(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"1.0", "1.0", 0},
		{"1.0", "2.0", -1},
		{"2.0", "1.0", 1},
	}

	for _, tc := range tests {
		a, _ := NewDecimalFromString(tc.a)
		b, _ := NewDecimalFromString(tc.b)
		result := a.Cmp(b)
		if result != tc.expected {
			t.Errorf("cmp(%s, %s): expected %d, got %d", tc.a, tc.b, tc.expected, result)
		}
	}
}

func TestDecimalRound(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		mode     RoundingMode
		decimals int
		expected string
	}{
		{"RoundHalfEven 1.235", "1.235", RoundHalfEven, 2, "1.24"},
		{"RoundHalfEven 1.225", "1.225", RoundHalfEven, 2, "1.22"},
		{"RoundUp 1.001", "1.001", RoundUp, 2, "1.01"},
		{"RoundDown 1.999", "1.999", RoundDown, 2, "1.99"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := NewDecimalFromString(tc.input)
			result := d.Round(tc.mode, tc.decimals)
			if result.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result.String())
			}
		})
	}
}

// Property-based tests
func TestProperty_AddSubIdentity(t *testing.T) {
	// m + 0 = m
	m := USD.FromInt(1000)
	zero := USD.Zero()

	result, err := m.Add(zero)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Amount() != m.Amount() {
		t.Errorf("m + 0 should equal m")
	}
}

func TestProperty_AddCommutative(t *testing.T) {
	a := USD.FromInt(500)
	b := USD.FromInt(300)

	resultAB, _ := a.Add(b)
	resultBA, _ := b.Add(a)

	if resultAB.Amount() != resultBA.Amount() {
		t.Errorf("a + b should equal b + a")
	}
}

func TestProperty_NegationIdentity(t *testing.T) {
	m := USD.FromInt(1000)
	neg := m.Neg()
	result, _ := m.Add(neg)
	zero := USD.Zero()

	if result.Amount() != zero.Amount() {
		t.Errorf("m + (-m) should equal 0")
	}
}

func TestProperty_SplitSum(t *testing.T) {
	m := USD.FromInt(1000)
	n := 3

	parts, err := m.Split(n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sum int64
	for _, p := range parts {
		sum += p.Amount()
	}

	if sum != m.Amount() {
		t.Errorf("sum of split parts should equal original: %d vs %d", sum, m.Amount())
	}
}

func TestProperty_AllocationSum(t *testing.T) {
	m := USD.FromInt(10000)
	ratios := []int{5, 3, 2}

	parts, err := m.AllocateRatios(ratios)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sum int64
	for _, p := range parts {
		sum += p.Amount()
	}

	if sum != m.Amount() {
		t.Errorf("sum of allocated parts should equal original: %d vs %d", sum, m.Amount())
	}
}

func TestProperty_AbsNonNegative(t *testing.T) {
	m := USD.FromInt(-1000)
	abs := m.Abs()

	if abs.Amount() < 0 {
		t.Errorf("abs(m) should be non-negative")
	}
}

func TestMoneyMulDecimal(t *testing.T) {
	// Test Money * Decimal operation
	m := USD.FromInt(1000)                      // $10.00
	factor, _ := NewDecimalFromString("1.0775") // tax rate
	result, err := m.Mul(factor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// $10.00 * 1.0775 = $10.775, rounded to $10.78 with RoundHalfEven
	if result.Amount() != 1078 {
		t.Errorf("expected 1078, got %d", result.Amount())
	}
}
