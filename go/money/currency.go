// Package money provides a decimal-based money arithmetic library for financial calculations.
//
// This package implements precise monetary arithmetic using integer-based storage
// in the currency's smallest unit (e.g., cents for USD). All operations maintain
// exact precision without the rounding errors inherent in floating-point arithmetic.
//
// # Key Features
//
//   - Integer-based storage: Amounts stored as int64 in smallest currency unit
//   - Currency-aware operations: All arithmetic validates currency matching
//   - Comprehensive rounding modes: 7 different strategies for precision reduction
//   - Error types for programmatic handling of edge cases
//
// # Quick Start
//
//	money.USD.FromString("29.99")        // Create from string
//	money.USD.FromInt(2999)              // Create from cents
//	price.Add(tax)                        // Add same-currency values
//	price.Multiply(rate, money.RoundHalfUp) // Multiply with rounding
//
// # Currency Codes
//
// The package includes pre-defined currencies following ISO 4217:
// USD, EUR, GBP, JPY, CHF, CAD, AUD, CNY, INR, BRL, MXN, KRW, SGD, HKD, NOK, SEK, DKK, NZD, ZAR
package money

import (
	"strings"
)

// Currency represents a currency with ISO 4217 properties.
// The Currency type is used to validate monetary operations and ensure
// that calculations only occur between values of the same currency.
// Currency values are typically used as a receiver for construction methods
// (e.g., USD.FromString("100.00")) and to specify the currency of Money values.
type Currency struct {
	code     string // ISO 4217 code: "USD", "EUR", etc.
	exponent int    // Number of decimal places: 2 for USD, 0 for JPY
	name     string // Human-readable name
}

// String returns the currency code.
func (c Currency) String() string {
	return c.code
}

// Code returns the ISO 4217 currency code.
func (c Currency) Code() string {
	return c.code
}

// DecimalPlaces returns the number of decimal places for this currency.
func (c Currency) DecimalPlaces() int {
	return c.exponent
}

// Name returns the full name of the currency.
func (c Currency) Name() string {
	return c.name
}

// Equal reports whether c and other represent the same currency.
func (c Currency) Equal(other Currency) bool {
	return c.code == other.code
}

// Predefined currencies
var (
	USD = Currency{code: "USD", exponent: 2, name: "United States Dollar"}
	EUR = Currency{code: "EUR", exponent: 2, name: "Euro"}
	GBP = Currency{code: "GBP", exponent: 2, name: "British Pound Sterling"}
	JPY = Currency{code: "JPY", exponent: 0, name: "Japanese Yen"}
	CHF = Currency{code: "CHF", exponent: 2, name: "Swiss Franc"}
	CAD = Currency{code: "CAD", exponent: 2, name: "Canadian Dollar"}
	AUD = Currency{code: "AUD", exponent: 2, name: "Australian Dollar"}
	CNY = Currency{code: "CNY", exponent: 2, name: "Chinese Yuan"}
	INR = Currency{code: "INR", exponent: 2, name: "Indian Rupee"}
	BRL = Currency{code: "BRL", exponent: 2, name: "Brazilian Real"}
	MXN = Currency{code: "MXN", exponent: 2, name: "Mexican Peso"}
	KRW = Currency{code: "KRW", exponent: 0, name: "South Korean Won"}
	SGD = Currency{code: "SGD", exponent: 2, name: "Singapore Dollar"}
	HKD = Currency{code: "HKD", exponent: 2, name: "Hong Kong Dollar"}
	NOK = Currency{code: "NOK", exponent: 2, name: "Norwegian Krone"}
	SEK = Currency{code: "SEK", exponent: 2, name: "Swedish Krona"}
	DKK = Currency{code: "DKK", exponent: 2, name: "Danish Krone"}
	NZD = Currency{code: "NZD", exponent: 2, name: "New Zealand Dollar"}
	ZAR = Currency{code: "ZAR", exponent: 2, name: "South African Rand"}
)

var currencyByCode = map[string]Currency{
	"USD": USD,
	"EUR": EUR,
	"GBP": GBP,
	"JPY": JPY,
	"CHF": CHF,
	"CAD": CAD,
	"AUD": AUD,
	"CNY": CNY,
	"INR": INR,
	"BRL": BRL,
	"MXN": MXN,
	"KRW": KRW,
	"SGD": SGD,
	"HKD": HKD,
	"NOK": NOK,
	"SEK": SEK,
	"DKK": DKK,
	"NZD": NZD,
	"ZAR": ZAR,
}

// ParseCurrency validates and returns a Currency for the given ISO 4217 code.
func ParseCurrency(code string) (Currency, error) {
	if len(code) != 3 {
		return Currency{}, InvalidCurrencyCodeError{CurrencyCode: code}
	}

	upper := strings.ToUpper(code)

	if currency, ok := currencyByCode[upper]; ok {
		return currency, nil
	}

	return Currency{}, UnknownCurrencyError{CurrencyCode: code}
}

// RegisteredCurrencies returns a list of all registered currency codes.
func RegisteredCurrencies() []string {
	codes := make([]string, 0, len(currencyByCode))
	for code := range currencyByCode {
		codes = append(codes, code)
	}
	return codes
}

// Scale returns the number of decimal places for this currency.
func (c Currency) Scale() int {
	return c.exponent
}
