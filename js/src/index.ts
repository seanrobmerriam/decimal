/**
 * Decimal Money Library - JavaScript/TypeScript API
 * 
 * This is a wrapper around the Rust/WebAssembly implementation.
 * For now, this provides a pure JS implementation for compatibility.
 * 
 * @module @decimal/money
 * @example
 * import { Money, Decimal, RoundingMode } from '@decimal/money';
 * 
 * const price = Money.fromString("USD", "29.99");
 * const tax = Decimal.fromString("0.08875");
 * const total = price.mul(tax);
 */

export { Money, Currency, RoundingMode, Decimal } from './money';

// Re-export types for external use
export type { RoundingModeType } from './money';
