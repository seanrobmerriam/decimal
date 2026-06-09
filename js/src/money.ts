/**
 * Decimal Money Library - Pure JavaScript Implementation
 * 
 * This provides a pure JS implementation for environments where
 * WebAssembly is not available. The API matches the Wasm version.
 * 
 * @module money
 */

/**
 * Rounding modes for monetary calculations.
 * Each mode defines how to handle precision reduction.
 */
export enum RoundingMode {
  /** Round values away from zero. Example: 1.001 -> 1.01, -1.001 -> -1.01 */
  Up = "UP",
  /** Round values toward zero. Example: 1.019 -> 1.01, -1.019 -> -1.01 */
  Down = "DOWN",
  /** Round up when discarded fraction >= 0.5. Example: 1.015 -> 1.02 */
  HalfUp = "HALF_UP",
  /** Round up only when discarded fraction > 0.5. Example: 1.015 -> 1.01 */
  HalfDown = "HALF_DOWN",
  /** Round to nearest even on tie. Banker's rounding. Example: 1.015 -> 1.02, 1.025 -> 1.02 */
  HalfEven = "HALF_EVEN",
  /** Round toward positive infinity. Example: 1.001 -> 1.01, -1.001 -> -1.00 */
  Ceiling = "CEILING",
  /** Round toward negative infinity. Example: 1.001 -> 1.00, -1.001 -> -1.01 */
  Floor = "FLOOR"
}

/**
 * Type alias for RoundingMode for backward compatibility.
 */
export type RoundingModeType = RoundingMode;

// Currency codes
export const CurrencyCode = {
  USD: { code: "USD", exponent: 2, name: "United States Dollar" },
  EUR: { code: "EUR", exponent: 2, name: "Euro" },
  GBP: { code: "GBP", exponent: 2, name: "British Pound Sterling" },
  JPY: { code: "JPY", exponent: 0, name: "Japanese Yen" },
  CHF: { code: "CHF", exponent: 2, name: "Swiss Franc" },
  CAD: { code: "CAD", exponent: 2, name: "Canadian Dollar" },
  AUD: { code: "AUD", exponent: 2, name: "Australian Dollar" },
  CNY: { code: "CNY", exponent: 2, name: "Chinese Yuan" },
  INR: { code: "INR", exponent: 2, name: "Indian Rupee" },
  BRL: { code: "BRL", exponent: 2, name: "Brazilian Real" },
  MXN: { code: "MXN", exponent: 2, name: "Mexican Peso" },
  KRW: { code: "KRW", exponent: 0, name: "South Korean Won" },
  SGD: { code: "SGD", exponent: 2, name: "Singapore Dollar" },
  HKD: { code: "HKD", exponent: 2, name: "Hong Kong Dollar" },
  NOK: { code: "NOK", exponent: 2, name: "Norwegian Krone" },
  SEK: { code: "SEK", exponent: 2, name: "Swedish Krona" },
  DKK: { code: "DKK", exponent: 2, name: "Danish Krone" },
  NZD: { code: "NZD", exponent: 2, name: "New Zealand Dollar" },
  ZAR: { code: "ZAR", exponent: 2, name: "South African Rand" }
} as const;

export type CurrencyType = typeof CurrencyCode.USD;

/**
 * Currency represents a currency with ISO 4217 properties.
 * Used to validate monetary operations and format Money values.
 */
export class Currency {
  /** ISO 4217 currency code (e.g., "USD", "EUR") */
  readonly code: string;
  /** Number of decimal places (2 for USD, 0 for JPY) */
  readonly exponent: number;
  /** Human-readable name (e.g., "United States Dollar") */
  readonly name: string;

  constructor(code: string, exponent: number, name: string) {
    this.code = code;
    this.exponent = exponent;
    this.name = name;
  }

  /**
   * Creates a Currency from an ISO 4217 code.
   * @param code - Three-letter currency code (case-insensitive)
   * @returns Currency instance or undefined if code is not recognized
   * 
   * @example
   * const usd = Currency.fromCode("USD");
   * const eur = Currency.fromCode("eur"); // case-insensitive
   */
  static fromCode(code: string): Currency | undefined {
    const c = CurrencyCode[code.toUpperCase() as keyof typeof CurrencyCode];
    if (c) {
      return new Currency(c.code, c.exponent, c.name);
    }
    return undefined;
  }
}

/**
 * Decimal represents an arbitrary-precision decimal number without currency association.
 * Used for intermediate calculations, percentages, and exchange rates.
 * 
 * The Decimal type uses a "bigint" unscaled value combined with a scale factor,
 * allowing it to represent values like "1.23" as (123, 2) where 123/10^2 = 1.23.
 */
export class Decimal {
  private value: bigint;
  private scale: number;

  constructor(value: bigint, scale: number) {
    this.value = value;
    this.scale = scale;
  }

  /**
   * Creates a Decimal from a string representation.
   * @param s - String representation of a decimal number
   * @throws Error if the string is empty or contains invalid characters
   * 
   * @example
   * const rate = Decimal.fromString("0.08875"); // 8.875%
   */
  static fromString(s: string): Decimal {
    s = s.trim();
    if (!s) {
      throw new Error("parse error: empty string");
    }

    const negative = s.startsWith('-');
    s = s.replace(/^[+-]/, '');

    const parts = s.split('.');
    const intPart = parts[0] || "0";
    const fracPart = parts[1] || "";

    if (!/^\d+$/.test(intPart) || !/^\d*$/.test(fracPart)) {
      throw new Error("parse error: invalid character");
    }

    const combined = intPart + fracPart;
    let value = BigInt(combined);
    if (negative) {
      value = -value;
    }

    return new Decimal(value, fracPart.length);
  }

  /** Returns the unscaled bigint value */
  getValue(): bigint {
    return this.value;
  }

  /** Returns the number of decimal places */
  getScale(): number {
    return this.scale;
  }

  /** Returns true if the value is zero */
  isZero(): boolean {
    return this.value === 0n;
  }

  /** Returns true if the value is greater than zero */
  isPositive(): boolean {
    return this.value > 0n;
  }

  /** Returns true if the value is less than zero */
  isNegative(): boolean {
    return this.value < 0n;
  }

  /** Returns the sum of two decimals */
  add(other: Decimal): Decimal {
    const scale = Math.max(this.scale, other.scale);
    const scaledThis = this.scaleTo(scale);
    const scaledOther = other.scaleTo(scale);
    return new Decimal(scaledThis.value + scaledOther.value, scale);
  }

  /** Returns the difference of two decimals (this - other) */
  sub(other: Decimal): Decimal {
    return this.add(other.neg());
  }

  /** Returns the product of two decimals */
  mul(other: Decimal): Decimal {
    return new Decimal(this.value * other.value, this.scale + other.scale);
  }

  /**
   * Returns the quotient of two decimals with the given rounding mode.
   * @param other - The divisor
   * @param rounding - Rounding mode (defaults to HalfEven)
   * @throws Error if divisor is zero
   */
  div(other: Decimal, rounding: RoundingMode = RoundingMode.HalfEven): Decimal {
    if (other.value === 0n) {
      throw new Error("division by zero");
    }

    const extraScale = 10;
    const scaledValue = this.value * 10n ** BigInt(extraScale);
    let quo = scaledValue / other.value;
    const rem = scaledValue % other.value;

    if (rem !== 0n) {
      const absRem = rem < 0n ? -rem : rem;
      const absDiv = other.value < 0n ? -other.value : other.value;
      const twoRem = absRem * 2n;

      let needsInc = false;
      switch (rounding) {
        case RoundingMode.Up:
          needsInc = true;
          break;
        case RoundingMode.Down:
          needsInc = false;
          break;
        case RoundingMode.HalfUp:
          needsInc = twoRem >= absDiv;
          break;
        case RoundingMode.HalfDown:
          needsInc = twoRem > absDiv;
          break;
        case RoundingMode.HalfEven:
          needsInc = twoRem > absDiv || (twoRem === absDiv && (quo < 0n ? -quo : quo) % 2n !== 0n);
          break;
        case RoundingMode.Ceiling:
          needsInc = quo >= 0n;
          break;
        case RoundingMode.Floor:
          needsInc = quo < 0n;
          break;
      }

      if (needsInc) {
        quo += quo >= 0n ? 1n : -1n;
      }
    }

    const scaleDivisor = 10n ** BigInt(extraScale);
    quo = quo / scaleDivisor;

    return new Decimal(quo, this.scale);
  }

  /**
   * Rounds the decimal to the specified number of decimal places.
   * @param rounding - Rounding mode to use
   * @param decimals - Target number of decimal places
   */
  round(rounding: RoundingMode, decimals: number): Decimal {
    if (decimals >= this.scale) {
      return this;
    }

    const diff = this.scale - decimals;
    const divisor = 10n ** BigInt(diff);
    let quo = this.value / divisor;
    const rem = this.value % divisor;

    if (rem !== 0n) {
      const absRem = rem < 0n ? -rem : rem;
      const absDiv = divisor;
      const twoRem = absRem * 2n;

      let needsInc = false;
      switch (rounding) {
        case RoundingMode.Up:
          needsInc = true;
          break;
        case RoundingMode.Down:
          needsInc = false;
          break;
        case RoundingMode.HalfUp:
          needsInc = twoRem >= absDiv;
          break;
        case RoundingMode.HalfDown:
          needsInc = twoRem > absDiv;
          break;
        case RoundingMode.HalfEven:
          needsInc = twoRem > absDiv || (twoRem === absDiv && (quo < 0n ? -quo : quo) % 2n !== 0n);
          break;
        case RoundingMode.Ceiling:
          needsInc = quo >= 0n;
          break;
        case RoundingMode.Floor:
          needsInc = quo < 0n;
          break;
      }

      if (needsInc) {
        quo += quo >= 0n ? 1n : -1n;
      }
    }

    return new Decimal(quo, decimals);
  }

  /** Returns the string representation of the decimal */
  toString(): string {
    const absValue = this.value < 0n ? -this.value : this.value;
    let s = absValue.toString();

    if (this.scale > 0) {
      if (s.length <= this.scale) {
        s = '0'.repeat(this.scale - s.length + 1) + s;
      }
      const pos = s.length - this.scale;
      s = s.slice(0, pos) + '.' + s.slice(pos);
    }

    return (this.value < 0n ? '-' : '') + s;
  }

  private scaleTo(targetScale: number): Decimal {
    if (targetScale <= this.scale) {
      return this;
    }
    const diff = targetScale - this.scale;
    return new Decimal(this.value * 10n ** BigInt(diff), targetScale);
  }

  private neg(): Decimal {
    return new Decimal(-this.value, this.scale);
  }
}

/**
 * Money represents a monetary amount with a specific currency.
 * All amounts are stored as integers in the currency's smallest unit (e.g., cents for USD).
 */
export class Money {
  private amount: bigint;
  private currency: Currency;

  constructor(amount: bigint, currency: Currency) {
    this.amount = amount;
    this.currency = currency;
  }

  /**
   * Creates Money from an integer amount in the currency's smallest unit.
   * @param currencyCode - ISO 4217 currency code (e.g., "USD")
   * @param amount - Amount in smallest unit (e.g., 1999 for $19.99)
   * @throws Error if currency code is not recognized
   * 
   * @example
   * const price = Money.fromInt("USD", 2999); // $29.99
   */
  static fromInt(currencyCode: string, amount: number | bigint): Money {
    const currency = Currency.fromCode(currencyCode);
    if (!currency) {
      throw new Error(`unknown currency: ${currencyCode}`);
    }
    return new Money(BigInt(amount), currency);
  }

  /**
   * Creates Money from a string representation.
   * @param currencyCode - ISO 4217 currency code
   * @param s - String representation (e.g., "29.99")
   * @throws Error if currency code is not recognized or string is invalid
   * 
   * @example
   * const price = Money.fromString("USD", "29.99"); // $29.99
   */
  static fromString(currencyCode: string, s: string): Money {
    const currency = Currency.fromCode(currencyCode);
    if (!currency) {
      throw new Error(`unknown currency: ${currencyCode}`);
    }

    s = s.trim();
    const negative = s.startsWith('-');
    s = s.replace(/^[+-]/, '');

    const parts = s.split('.');
    const intPart = parts[0] || "0";
    const fracPart = parts[1] || "";

    if (!/^\d+$/.test(intPart) || !/^\d*$/.test(fracPart)) {
      throw new Error("parse error: invalid character");
    }

    let frac = fracPart;
    while (frac.length < currency.exponent) {
      frac += '0';
    }
    if (frac.length > currency.exponent) {
      frac = frac.slice(0, currency.exponent);
    }

    const combined = intPart + frac;
    let value = BigInt(combined);
    if (negative) {
      value = -value;
    }

    return new Money(value, currency);
  }

  /** Returns the amount in the smallest currency unit */
  getAmount(): bigint {
    return this.amount;
  }

  /** Returns the currency */
  getCurrency(): Currency {
    return this.currency;
  }

  /** Returns true if the amount is zero */
  isZero(): boolean {
    return this.amount === 0n;
  }

  /** Returns true if the amount is greater than zero */
  isPositive(): boolean {
    return this.amount > 0n;
  }

  /** Returns true if the amount is less than zero */
  isNegative(): boolean {
    return this.amount < 0n;
  }

  /**
   * Adds another Money value. Both must have the same currency.
   * @param other - Money to add
   * @throws Error if currencies don't match
   */
  add(other: Money): Money {
    if (this.currency.code !== other.currency.code) {
      throw new Error(`currency mismatch: cannot add ${this.currency.code} and ${other.currency.code}`);
    }

    const result = this.amount + other.amount;
    return new Money(result, this.currency);
  }

  /**
   * Subtracts another Money value. Both must have the same currency.
   * @param other - Money to subtract
   * @throws Error if currencies don't match
   */
  sub(other: Money): Money {
    if (this.currency.code !== other.currency.code) {
      throw new Error(`currency mismatch: cannot subtract ${this.currency.code} and ${other.currency.code}`);
    }

    const result = this.amount - other.amount;
    return new Money(result, this.currency);
  }

  /**
   * Multiplies the money by a decimal factor.
   * @param factor - Decimal factor to multiply by
   * @returns Result rounded to currency's decimal places
   */
  mul(factor: Decimal): Money {
    const product = this.amount * factor.getValue();
    const scale = factor.getScale();

    if (scale > 0) {
      const divisor = 10n ** BigInt(scale);
      let quo = product / divisor;
      const rem = product % divisor;

      if (rem !== 0n) {
        const absRem = rem < 0n ? -rem : rem;
        const absDiv = divisor;
        const twoRem = absRem * 2n;

        const needsInc = twoRem >= absDiv;
        if (needsInc) {
          quo += product > 0n ? 1n : -1n;
        }
      }

      return new Money(quo, this.currency);
    }

    return new Money(product, this.currency);
  }

  /**
   * Divides the money by a decimal divisor.
   * @param divisor - Decimal divisor
   * @param rounding - Rounding mode (defaults to HalfEven)
   * @throws Error if divisor is zero
   */
  div(divisor: Decimal, rounding: RoundingMode = RoundingMode.HalfEven): Money {
    if (divisor.getValue() === 0n) {
      throw new Error("division by zero");
    }

    const extraScale = 4;
    const scaled = this.amount * 10n ** BigInt(extraScale);
    let quo = scaled / divisor.getValue();
    const rem = scaled % divisor.getValue();

    if (rem !== 0n) {
      const absRem = rem < 0n ? -rem : rem;
      const absDiv = divisor.getValue() < 0n ? -divisor.getValue() : divisor.getValue();
      const twoRem = absRem * 2n;

      let needsInc = false;
      switch (rounding) {
        case RoundingMode.Up:
          needsInc = true;
          break;
        case RoundingMode.Down:
          needsInc = false;
          break;
        case RoundingMode.HalfUp:
          needsInc = twoRem >= absDiv;
          break;
        case RoundingMode.HalfDown:
          needsInc = twoRem > absDiv;
          break;
        case RoundingMode.HalfEven:
          needsInc = twoRem > absDiv || (twoRem === absDiv && (quo < 0n ? -quo : quo) % 2n !== 0n);
          break;
        case RoundingMode.Ceiling:
          needsInc = quo >= 0n;
          break;
        case RoundingMode.Floor:
          needsInc = quo < 0n;
          break;
      }

      if (needsInc) {
        quo += quo >= 0n ? 1n : -1n;
      }
    }

    const scaleDivisor = 10n ** BigInt(extraScale);
    quo = quo / scaleDivisor;

    return new Money(quo, this.currency);
  }

  split(n: number): Money[] {
    if (n <= 0) {
      throw new Error("n must be positive");
    }

    if (n === 1) {
      return [this];
    }

    const base = this.amount / BigInt(n);
    const remainder = this.amount % BigInt(n);

    const result: Money[] = [];
    for (let i = 0; i < n; i++) {
      let amount = base;
      if (remainder > 0n && BigInt(i) < remainder) {
        amount += 1n;
      } else if (remainder < 0n && BigInt(i) < -remainder) {
        amount -= 1n;
      }
      result.push(new Money(amount, this.currency));
    }

    return result;
  }

  allocateRatios(ratios: number[]): Money[] {
    if (ratios.length === 0) {
      throw new Error("ratios cannot be empty");
    }

    let totalRatio = ratios.reduce((sum, r) => sum + Math.max(0, r), 0);
    if (totalRatio === 0) {
      throw new Error("sum of ratios must be positive");
    }

    const result: Money[] = [];
    let remaining = this.amount;

    for (let i = 0; i < ratios.length; i++) {
      if (i === ratios.length - 1) {
        result.push(new Money(remaining, this.currency));
      } else {
        const share = remaining * BigInt(ratios[i]) / BigInt(totalRatio);
        result.push(new Money(share, this.currency));
        remaining -= share;
        totalRatio -= ratios[i];
      }
    }

    return result;
  }

  toString(): string {
    const exp = this.currency.exponent;
    const absAmount = this.amount < 0n ? -this.amount : this.amount;

    let amountStr = absAmount.toString();
    if (exp > 0) {
      if (amountStr.length <= exp) {
        amountStr = '0'.repeat(exp - amountStr.length + 1) + amountStr;
      }
      const pos = amountStr.length - exp;
      amountStr = amountStr.slice(0, pos) + '.' + amountStr.slice(pos);
    }

    if (this.amount < 0n) {
      amountStr = '-' + amountStr;
    }

    return `${this.currency.code} ${amountStr}`;
  }

  format(): string {
    return this.toString();
  }

  compareTo(other: Money): -1 | 0 | 1 {
    if (this.currency.code !== other.currency.code) {
      throw new Error(`currency mismatch`);
    }
    if (this.amount < other.amount) return -1;
    if (this.amount > other.amount) return 1;
    return 0;
  }
}
