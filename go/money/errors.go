package money

import (
	"fmt"
)

// ErrorKind categorizes errors for programmatic handling.
// Use the Kind() method on any Error to determine the category.
type ErrorKind int

// Error kind constants identify the category of error.
// These are used for programmatic error handling via a type switch.
const (
	// ErrKindCurrencyMismatch indicates operations were attempted between Money
	// values of different currencies.
	ErrKindCurrencyMismatch ErrorKind = iota

	// ErrKindDivisionByZero indicates division by zero was attempted.
	ErrKindDivisionByZero

	// ErrKindOverflow indicates a calculation result exceeds int64 range.
	ErrKindOverflow

	// ErrKindUnderflow indicates a calculation produced a value too small
	// to represent at the required precision.
	ErrKindUnderflow

	// ErrKindPrecisionLoss indicates precision was lost during a conversion.
	// Typically from float64 to integer-based money.
	ErrKindPrecisionLoss

	// ErrKindParse indicates a string could not be parsed as money/decimal.
	ErrKindParse

	// ErrKindInvalidOperation indicates an invalid operation was attempted.
	ErrKindInvalidOperation

	// ErrKindRounding indicates rounding produced unexpected results.
	ErrKindRounding
)

// Error codes for machine-readable error handling.
// These string codes can be used in switch statements or for API responses.
const (
	CodeCurrencyMismatch = "CURRENCY_MISMATCH" // Currencies don't match in operation
	CodeDivisionByZero   = "DIVISION_BY_ZERO"  // Division by zero attempted
	CodeOverflow         = "OVERFLOW"          // Result exceeds int64 range
	CodeUnderflow        = "UNDERFLOW"         // Value too small for precision
	CodePrecisionLoss    = "PRECISION_LOSS"    // Lost digits in conversion
	CodeParseError       = "PARSE_ERROR"       // String parsing failed
	CodeInvalidOperation = "INVALID_OPERATION" // Operation not valid
	CodeRounding         = "ROUNDING_ERROR"    // Rounding issue
	CodeInvalidCurrency  = "INVALID_CURRENCY"  // Currency code invalid format
	CodeUnknownCurrency  = "UNKNOWN_CURRENCY"  // Currency code not recognized
)

// Error represents an error in money operations.
// All errors returned by this package implement this interface,
// allowing for unified error handling and type switches.
type Error interface {
	error
	// Kind returns the ErrorKind category for programmatic handling.
	Kind() ErrorKind
	// Code returns a machine-readable string code for API responses.
	Code() string
	// Details returns additional context about the error.
	Details() map[string]interface{}
}

// CurrencyMismatchError is returned when operations are attempted
// between Money values of different currencies. Money arithmetic
// requires both operands to have the same currency.
//
// Example handling:
//
//	result, err := usdAmount.Add(eurAmount)
//	if err != nil {
//	    if merr, ok := err.(money.CurrencyMismatchError); ok {
//	        // Handle currency mismatch
//	    }
//	}
type CurrencyMismatchError struct {
	Left      Currency // Left operand currency
	Right     Currency // Right operand currency
	Operation string   // Operation name: "add", "sub", "mul", "div", "cmp"
}

func (e CurrencyMismatchError) Error() string {
	return fmt.Sprintf("currency mismatch: cannot %s %s (%d decimals) and %s (%d decimals)",
		e.Operation, e.Left.code, e.Left.exponent, e.Right.code, e.Right.exponent)
}

func (e CurrencyMismatchError) Kind() ErrorKind { return ErrKindCurrencyMismatch }
func (e CurrencyMismatchError) Code() string    { return CodeCurrencyMismatch }
func (e CurrencyMismatchError) Details() map[string]interface{} {
	return map[string]interface{}{
		"left_currency":  e.Left.code,
		"right_currency": e.Right.code,
		"operation":      e.Operation,
	}
}

// DivisionByZeroError is returned when dividing by zero.
type DivisionByZeroError struct {
	Dividend Money    // The value being divided
	Divisor  *Decimal // The zero divisor
	Context  string   // Where the error occurred
}

func (e DivisionByZeroError) Error() string {
	return fmt.Sprintf("division by zero: cannot divide %s by zero", e.Dividend.String())
}

func (e DivisionByZeroError) Kind() ErrorKind { return ErrKindDivisionByZero }
func (e DivisionByZeroError) Code() string    { return CodeDivisionByZero }
func (e DivisionByZeroError) Details() map[string]interface{} {
	return map[string]interface{}{
		"dividend": e.Dividend.String(),
		"context":  e.Context,
	}
}

// OverflowError is returned when a calculation result exceeds int64 range.
type OverflowError struct {
	Operation string                 // Operation that caused overflow
	Left      Money                  // Left operand
	Right     *Decimal               // Right operand
	MaxValue  int64                  // Maximum representable value
	Context   map[string]interface{} // Additional context
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("overflow: result exceeds maximum value %d from operation %s",
		e.MaxValue, e.Operation)
}

func (e OverflowError) Kind() ErrorKind { return ErrKindOverflow }
func (e OverflowError) Code() string    { return CodeOverflow }
func (e OverflowError) Details() map[string]interface{} {
	return e.Context
}

// UnderflowError is returned when a calculation produces a value
// too small to represent at the required precision.
type UnderflowError struct {
	Operation    string   // Operation that caused underflow
	Value        *Decimal // The value that underflowed
	MinPrecision int      // Minimum representable precision
	Context      string   // Where the error occurred
}

func (e UnderflowError) Error() string {
	return fmt.Sprintf("underflow: value %s below minimum representable at precision %d",
		e.Value.String(), e.MinPrecision)
}

func (e UnderflowError) Kind() ErrorKind { return ErrKindUnderflow }
func (e UnderflowError) Code() string    { return CodeUnderflow }

// PrecisionLossError is returned when precision is lost in a conversion.
type PrecisionLossError struct {
	Original   string // Original string representation
	Converted  string // Converted value (approximation)
	LostDigits int    // Number of significant digits lost
	Context    string // Where the loss occurred
}

func (e PrecisionLossError) Error() string {
	return fmt.Sprintf("precision loss: %s converted to %s, lost %d digits",
		e.Original, e.Converted, e.LostDigits)
}

func (e PrecisionLossError) Kind() ErrorKind { return ErrKindPrecisionLoss }
func (e PrecisionLossError) Code() string    { return CodePrecisionLoss }

// ParseError is returned when a string cannot be parsed.
type ParseError struct {
	Input    string // The string that failed to parse
	Position int    // Position of error (0 if unknown)
	Reason   string // Why parsing failed
	Expected string // What was expected
}

func (e ParseError) Error() string {
	if e.Position > 0 {
		return fmt.Sprintf("parse error at position %d: %s (expected %s)",
			e.Position, e.Reason, e.Expected)
	}
	return fmt.Sprintf("parse error: %s (expected %s)", e.Reason, e.Expected)
}

func (e ParseError) Kind() ErrorKind { return ErrKindParse }
func (e ParseError) Code() string    { return CodeParseError }
func (e ParseError) Details() map[string]interface{} {
	return map[string]interface{}{
		"input":    e.Input,
		"position": e.Position,
		"reason":   e.Reason,
		"expected": e.Expected,
	}
}

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

func (e InvalidOperationError) Kind() ErrorKind { return ErrKindInvalidOperation }
func (e InvalidOperationError) Code() string    { return CodeInvalidOperation }

// RoundingError is returned when rounding produces unexpected results.
type RoundingError struct {
	Original *Decimal     // Value before rounding
	Rounded  *Decimal     // Value after rounding
	Mode     RoundingMode // Rounding mode used
	Reason   string       // Why rounding was problematic
}

func (e RoundingError) Error() string {
	return fmt.Sprintf("rounding warning: %s -> %s using %s (%s)",
		e.Original.String(), e.Rounded.String(), e.Mode.String(), e.Reason)
}

func (e RoundingError) Kind() ErrorKind { return ErrKindRounding }
func (e RoundingError) Code() string    { return CodeRounding }

// InvalidCurrencyCodeError is returned when a currency code is invalid.
type InvalidCurrencyCodeError struct {
	CurrencyCode string
}

func (e InvalidCurrencyCodeError) Error() string {
	return fmt.Sprintf("invalid currency code: %s (must be 3 letters)", e.CurrencyCode)
}

func (e InvalidCurrencyCodeError) Kind() ErrorKind { return ErrKindInvalidOperation }
func (e InvalidCurrencyCodeError) Code() string    { return CodeInvalidCurrency }

// UnknownCurrencyError is returned when a currency code is not registered.
type UnknownCurrencyError struct {
	CurrencyCode string
}

func (e UnknownCurrencyError) Error() string {
	return fmt.Sprintf("unknown currency: %s", e.CurrencyCode)
}

func (e UnknownCurrencyError) Kind() ErrorKind { return ErrKindInvalidOperation }
func (e UnknownCurrencyError) Code() string    { return CodeUnknownCurrency }
