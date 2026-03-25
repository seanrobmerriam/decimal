package allocator

import (
	"testing"

	"github.com/decimal/money/money"
)

func TestAllocateEven(t *testing.T) {
	// Allocate $10 equally among 3 people
	total := money.USD.FromInt(1000) // $10.00
	result, err := Allocate(total, 3)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(result))
	}

	// Check first gets extra cent: $3.34, $3.33, $3.33
	expected := []int64{334, 333, 333}
	for i, part := range result {
		if part.Amount() != expected[i] {
			t.Errorf("Part %d: expected %d cents, got %d", i, expected[i], part.Amount())
		}
	}
}

func TestAllocateZero(t *testing.T) {
	total := money.USD.FromInt(0)
	result, err := Allocate(total, 3)
	if err != nil {
		t.Fatalf("Allocate with zero should not fail: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(result))
	}

	for i, part := range result {
		if part.Amount() != 0 {
			t.Errorf("Part %d: expected 0, got %d", i, part.Amount())
		}
	}
}

func TestAllocateNegative(t *testing.T) {
	total := money.USD.FromInt(-1000) // -$10.00
	result, err := Allocate(total, 3)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(result))
	}

	// Go integer division truncates toward zero for negatives
	// -1000 / 3 = -333, -1000 % 3 = -1
	// Since remainder is negative, no extra cent is given
	// Result: -$3.33, -$3.33, -$3.33
	expected := []int64{-333, -333, -333}
	for i, part := range result {
		if part.Amount() != expected[i] {
			t.Errorf("Part %d: expected %d cents, got %d", i, expected[i], part.Amount())
		}
	}
}

func TestAllocateByRatios(t *testing.T) {
	// Allocate $100 with ratios 5:3:2
	total := money.USD.FromInt(10000) // $100.00
	result, err := AllocateRatios(total, []int{5, 3, 2})
	if err != nil {
		t.Fatalf("AllocateRatios failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(result))
	}

	// 5+3+2=10 total parts, $10.00 each = $100.00
	// Part 0: 5 * 1000 = 5000 cents = $50.00
	// Part 1: 3 * 1000 = 3000 cents = $30.00
	// Part 2: 2 * 1000 = 2000 cents = $20.00
	expected := []int64{5000, 3000, 2000}
	for i, part := range result {
		if part.Amount() != expected[i] {
			t.Errorf("Part %d: expected %d cents, got %d", i, expected[i], part.Amount())
		}
	}
}

func TestAllocateByRatiosZeroTotal(t *testing.T) {
	total := money.USD.FromInt(0)
	result, err := AllocateRatios(total, []int{5, 3, 2})
	if err != nil {
		t.Fatalf("AllocateRatios with zero total should not fail: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 parts, got %d", len(result))
	}
}

func TestAllocateByRatiosEmptyRatios(t *testing.T) {
	total := money.USD.FromInt(10000)
	_, err := AllocateRatios(total, []int{})
	if err == nil {
		t.Error("Expected error for empty ratios")
	}
}

func TestAllocateByRatiosNegativeRatios(t *testing.T) {
	total := money.USD.FromInt(10000)
	_, err := AllocateRatios(total, []int{5, -3, 2})
	if err == nil {
		t.Error("Expected error for negative ratios")
	}
}

func TestAllocateByRatiosSumPreservation(t *testing.T) {
	// Property test: allocated parts should sum to original
	total := money.USD.FromInt(10007)                    // $100.07 with remainder
	result, err := AllocateRatios(total, []int{1, 1, 1}) // Equal thirds
	if err != nil {
		t.Fatalf("AllocateRatios failed: %v", err)
	}

	var sum int64
	for _, part := range result {
		sum += part.Amount()
	}

	if sum != total.Amount() {
		t.Errorf("Sum of parts (%d) should equal original (%d)", sum, total.Amount())
	}
}

func TestAllocateInvalidN(t *testing.T) {
	total := money.USD.FromInt(1000)
	_, err := Allocate(total, 0)
	if err == nil {
		t.Error("Expected error for n=0")
	}

	_, err = Allocate(total, -1)
	if err == nil {
		t.Error("Expected error for n=-1")
	}
}

func TestAllocateSingleRecipient(t *testing.T) {
	total := money.USD.FromInt(1000)
	result, err := Allocate(total, 1)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(result))
	}

	if result[0].Amount() != total.Amount() {
		t.Errorf("Single allocation should equal original: got %d, want %d", result[0].Amount(), total.Amount())
	}
}

func TestAllocateRatiosSingleRatio(t *testing.T) {
	total := money.USD.FromInt(1000)
	result, err := AllocateRatios(total, []int{1})
	if err != nil {
		t.Fatalf("AllocateRatios failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(result))
	}

	if result[0].Amount() != total.Amount() {
		t.Errorf("Single ratio should get all: got %d, want %d", result[0].Amount(), total.Amount())
	}
}
