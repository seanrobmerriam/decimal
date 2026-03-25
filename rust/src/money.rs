use wasm_bindgen::prelude::*;

use crate::decimal::Decimal;
use crate::errors::Error;

#[wasm_bindgen]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RoundingMode {
    Up,
    Down,
    HalfUp,
    HalfDown,
    HalfEven,
    Ceiling,
    Floor,
}

impl Default for RoundingMode {
    fn default() -> Self {
        RoundingMode::HalfEven
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Currency {
    code: &'static str,
    exponent: i32,
    name: &'static str,
}

#[wasm_bindgen]
impl Currency {
    #[wasm_bindgen(getter)]
    pub fn code(&self) -> String {
        self.code.to_string()
    }

    #[wasm_bindgen(getter)]
    pub fn exponent(&self) -> i32 {
        self.exponent
    }

    #[wasm_bindgen(getter)]
    pub fn name(&self) -> String {
        self.name.to_string()
    }
}

impl Currency {
    pub fn new(code: &'static str, exponent: i32, name: &'static str) -> Currency {
        Currency { code, exponent, name }
    }
}

macro_rules! define_currencies {
    ($($code:ident: $name:expr, $exp:expr, $full:expr);*) => {
        $(
            pub const $code: Currency = Currency {
                code: $name,
                exponent: $exp,
                name: $full,
            };
        )*

        pub fn get_currency(code: &str) -> Option<Currency> {
            match code.to_uppercase().as_str() {
                $(stringify!($code) => Some($code),)*
                _ => None,
            }
        }
    }
}

define_currencies!(
    USD: "USD", 2, "United States Dollar";
    EUR: "EUR", 2, "Euro";
    GBP: "GBP", 2, "British Pound Sterling";
    JPY: "JPY", 0, "Japanese Yen";
    CHF: "CHF", 2, "Swiss Franc";
    CAD: "CAD", 2, "Canadian Dollar";
    AUD: "AUD", 2, "Australian Dollar";
    CNY: "CNY", 2, "Chinese Yuan";
    INR: "INR", 2, "Indian Rupee";
    BRL: "BRL", 2, "Brazilian Real";
    MXN: "MXN", 2, "Mexican Peso";
    KRW: "KRW", 0, "South Korean Won";
    SGD: "SGD", 2, "Singapore Dollar";
    HKD: "HKD", 2, "Hong Kong Dollar";
    NOK: "NOK", 2, "Norwegian Krone";
    SEK: "SEK", 2, "Swedish Krona";
    DKK: "DKK", 2, "Danish Krone";
    NZD: "NZD", 2, "New Zealand Dollar";
    ZAR: "ZAR", 2, "South African Rand"
);

#[wasm_bindgen]
#[derive(Debug, Clone)]
pub struct Money {
    amount: i64,
    currency: Currency,
}

#[wasm_bindgen]
impl Money {
    #[wasm_bindgen(constructor)]
    pub fn from_int(currency_code: &str, amount: i64) -> Result<Money, Error> {
        let currency = get_currency(currency_code)
            .ok_or_else(|| Error::new_invalid_operation(&format!("unknown currency: {}", currency_code)))?;
        Ok(Money { amount, currency })
    }

    pub fn from_string(currency_code: &str, s: &str) -> Result<Money, Error> {
        let currency = get_currency(currency_code)
            .ok_or_else(|| Error::new_invalid_operation(&format!("unknown currency: {}", currency_code)))?;
        
        let s = s.trim();
        let negative = s.starts_with('-');
        let s = s.trim_start_matches('-').trim_start_matches('+');

        let parts: Vec<&str> = s.split('.').collect();
        let int_part = parts.get(0).unwrap_or(&"0");
        let frac_part = parts.get(1).unwrap_or(&"");

        if !int_part.chars().all(|c| c.is_ascii_digit()) {
            return Err(Error::new_parse_error(s, "invalid integer part"));
        }

        let mut frac = frac_part.to_string();
        while frac.len() < currency.exponent as usize {
            frac.push('0');
        }
        if frac.len() > currency.exponent as usize {
            frac = frac[..currency.exponent as usize].to_string();
        }

        let combined = format!("{}{}", int_part, frac);
        let value: i64 = combined.parse().map_err(|_| Error::new_parse_error(s, "invalid number"))?;

        let value = if negative { -value } else { value };

        Ok(Money { amount: value, currency })
    }

    #[wasm_bindgen(getter)]
    pub fn amount(&self) -> i64 {
        self.amount
    }

    #[wasm_bindgen(getter)]
    pub fn currency_code(&self) -> String {
        self.currency.code.to_string()
    }

    pub fn is_zero(&self) -> bool {
        self.amount == 0
    }

    pub fn is_positive(&self) -> bool {
        self.amount > 0
    }

    pub fn is_negative(&self) -> bool {
        self.amount < 0
    }
}

impl Money {
    pub fn new(amount: i64, currency: Currency) -> Money {
        Money { amount, currency }
    }

    pub fn currency(&self) -> Currency {
        self.currency
    }

    pub fn add(&self, other: &Money) -> Result<Money, Error> {
        if self.currency != other.currency {
            return Err(Error::new_currency_mismatch(
                self.currency.code,
                other.currency.code,
                "add",
            ));
        }

        let result = self.amount.checked_add(other.amount)
            .ok_or_else(|| Error::new_overflow("add"))?;

        Ok(Money {
            amount: result,
            currency: self.currency,
        })
    }

    pub fn sub(&self, other: &Money) -> Result<Money, Error> {
        if self.currency != other.currency {
            return Err(Error::new_currency_mismatch(
                self.currency.code,
                other.currency.code,
                "subtract",
            ));
        }

        let result = self.amount.checked_sub(other.amount)
            .ok_or_else(|| Error::new_overflow("subtract"))?;

        Ok(Money {
            amount: result,
            currency: self.currency,
        })
    }

    pub fn mul(&self, factor: &Decimal) -> Result<Money, Error> {
        let product = self.amount * factor.value;
        let scale = factor.scale;

        if scale > 0 {
            let divisor = 10_i64.pow(scale as u32);
            let quo = product / divisor;
            let rem = product % divisor;

            let mut result = quo;
            if rem != 0 {
                let abs_rem = rem.abs();
                let abs_div = divisor;
                let two_rem = abs_rem * 2;

                let needs_inc = two_rem >= abs_div;
                if needs_inc {
                    result += if product > 0 { 1 } else { -1 };
                }
            }

            Ok(Money {
                amount: result,
                currency: self.currency,
            })
        } else {
            Ok(Money {
                amount: product,
                currency: self.currency,
            })
        }
    }

    pub fn div(&self, divisor: &Decimal, rounding: RoundingMode) -> Result<Money, Error> {
        if divisor.value == 0 {
            return Err(Error::new_division_by_zero());
        }

        let extra_scale = 4;
        let scaled = self.amount * 10_i64.pow(extra_scale);
        let quo = scaled / divisor.value;
        let rem = scaled % divisor.value;

        let mut result = quo;
        if rem != 0 {
            let abs_rem = rem.abs();
            let abs_div = divisor.value.abs();
            let two_rem = abs_rem * 2;

            let needs_inc = match rounding {
                RoundingMode::Up => true,
                RoundingMode::Down => false,
                RoundingMode::HalfUp => two_rem >= abs_div,
                RoundingMode::HalfDown => two_rem > abs_div,
                RoundingMode::HalfEven => {
                    if two_rem > abs_div {
                        true
                    } else if two_rem == abs_div && quo.abs() % 2 != 0 {
                        true
                    } else {
                        false
                    }
                }
                RoundingMode::Ceiling => self.amount > 0 && rem != 0,
                RoundingMode::Floor => self.amount < 0 && rem != 0,
            };

            if needs_inc {
                result += if divisor.value > 0 { 1 } else { -1 };
            }
        }

        Ok(Money {
            amount: result,
            currency: self.currency,
        })
    }

    pub fn split(&self, n: u32) -> Result<Vec<Money>, Error> {
        if n == 0 {
            return Err(Error::new_invalid_operation("n must be positive"));
        }

        if n == 1 {
            return Ok(vec![self.clone()]);
        }

        let base = self.amount / n as i64;
        let remainder = self.amount % n as i64;

        let mut result = Vec::with_capacity(n as usize);
        for i in 0..n {
            let amount = base + if (i as i64) < remainder { 1 } else { 0 };
            result.push(Money {
                amount,
                currency: self.currency,
            });
        }

        Ok(result)
    }

    pub fn to_string(&self) -> String {
        let exp = self.currency.exponent;
        let abs_amount = self.amount.abs();
        
        let amount_str = if exp == 0 {
            abs_amount.to_string()
        } else {
            let s = abs_amount.to_string();
            if s.len() <= exp as usize {
                format!("0{}", s)
            } else {
                let pos = s.len() - exp as usize;
                format!("{}.{}", &s[..pos], &s[pos..])
            }
        };

        let result = if self.amount < 0 {
            format!("-{}", amount_str)
        } else {
            amount_str
        };

        format!("{} {}", self.currency.code, result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_money_add() {
        let a = Money::from_string("USD", "10.00").unwrap();
        let b = Money::from_string("USD", "5.00").unwrap();
        let result = a.add(&b).unwrap();
        assert_eq!(result.amount, 1500);
    }

    #[test]
    fn test_money_split() {
        let m = Money::from_string("USD", "10.00").unwrap();
        let parts = m.split(3).unwrap();
        assert_eq!(parts.len(), 3);
        assert_eq!(parts[0].amount, 334);
        assert_eq!(parts[1].amount, 333);
        assert_eq!(parts[2].amount, 333);
    }
}
