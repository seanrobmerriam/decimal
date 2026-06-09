use wasm_bindgen::prelude::*;

use crate::errors::Error;

#[wasm_bindgen]
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Decimal {
    value: i64,
    scale: i32,
}

#[wasm_bindgen]
impl Decimal {
    #[wasm_bindgen(constructor)]
    pub fn from_string(s: &str) -> Result<Decimal, Error> {
        let s = s.trim();
        if s.is_empty() {
            return Err(Error::new_parse_error(s, "empty string"));
        }

        let negative = s.starts_with('-');
        let s = s.trim_start_matches('-').trim_start_matches('+');

        let parts: Vec<&str> = s.split('.').collect();
        let int_part = parts.get(0).unwrap_or(&"0");
        let frac_part = parts.get(1).unwrap_or(&"");

        if !int_part.chars().all(|c| c.is_ascii_digit()) {
            return Err(Error::new_parse_error(s, "invalid integer part"));
        }

        if !frac_part.chars().all(|c| c.is_ascii_digit()) {
            return Err(Error::new_parse_error(s, "invalid fractional part"));
        }

        let combined = format!("{}{}", int_part, frac_part);
        let value: i128 = combined.parse().map_err(|_| Error::new_parse_error(s, "invalid number"))?;

        let value = if negative { -value } else { value };
        if value > i64::MAX as i128 || value < i64::MIN as i128 {
            return Err(Error::new_parse_error(s, "value out of range"));
        }
        let value = value as i64;
        let scale = frac_part.len() as i32;

        Ok(Decimal { value, scale })
    }

    #[wasm_bindgen(getter)]
    pub fn value(&self) -> i64 {
        self.value
    }

    #[wasm_bindgen(getter)]
    pub fn scale(&self) -> i32 {
        self.scale
    }

    pub fn is_zero(&self) -> bool {
        self.value == 0
    }

    pub fn is_positive(&self) -> bool {
        self.value > 0
    }

    pub fn is_negative(&self) -> bool {
        self.value < 0
    }
}

impl Decimal {
    pub fn new(value: i64, scale: i32) -> Decimal {
        Decimal { value, scale }
    }

    pub fn from_i64(value: i64) -> Decimal {
        Decimal { value, scale: 0 }
    }

    pub fn add(&self, other: &Decimal) -> Decimal {
        let scale = self.scale.max(other.scale);
        let scaled_self = self.scale_to(scale);
        let scaled_other = other.scale_to(scale);
        Decimal {
            value: scaled_self.value + scaled_other.value,
            scale,
        }
    }

    pub fn sub(&self, other: &Decimal) -> Decimal {
        self.add(&other.neg())
    }

    pub fn mul(&self, other: &Decimal) -> Decimal {
        Decimal {
            value: self.value * other.value,
            scale: self.scale + other.scale,
        }
    }

    pub fn div(&self, other: &Decimal, rounding: crate::money::RoundingMode) -> Result<Decimal, Error> {
        if other.value == 0 {
            return Err(Error::new_division_by_zero());
        }

        let extra_scale = 10;
        let scaled_value = self.value * 10_i64.pow(extra_scale as u32);
        let quo = scaled_value / other.value;
        let rem = scaled_value % other.value;

        let mut scaled_result = quo;
        if rem != 0 {
            let abs_rem = rem.abs();
            let abs_div = other.value.abs();

            let needs_inc = match rounding {
                crate::money::RoundingMode::Up => true,
                crate::money::RoundingMode::Down => false,
                crate::money::RoundingMode::HalfUp => abs_rem * 2 >= abs_div,
                crate::money::RoundingMode::HalfDown => abs_rem * 2 > abs_div,
                crate::money::RoundingMode::HalfEven => {
                    abs_rem * 2 > abs_div || (abs_rem * 2 == abs_div && quo.abs() % 2 != 0)
                }
                crate::money::RoundingMode::Ceiling => {
                    scaled_result >= 0
                }
                crate::money::RoundingMode::Floor => {
                    scaled_result < 0
                }
            };

            if needs_inc {
                scaled_result += if scaled_result >= 0 { 1 } else { -1 };
            }
        }

        let scale_divisor = 10_i64.pow(extra_scale as u32);
        let result_value = scaled_result / scale_divisor;

        Ok(Decimal {
            value: result_value,
            scale: self.scale,
        })
    }

    fn scale_to(&self, target_scale: i32) -> Decimal {
        if target_scale <= self.scale {
            return self.clone();
        }
        let diff = target_scale - self.scale;
        Decimal {
            value: self.value * 10_i64.pow(diff as u32),
            scale: target_scale,
        }
    }

    fn neg(&self) -> Decimal {
        Decimal {
            value: self.value.wrapping_neg(),
            scale: self.scale,
        }
    }

    pub fn round(&self, rounding: crate::money::RoundingMode, decimals: i32) -> Decimal {
        if decimals >= self.scale {
            return self.clone();
        }

        let diff = self.scale - decimals;
        let divisor = 10_i64.pow(diff as u32);
        let quo = self.value / divisor;
        let rem = self.value % divisor;

        let mut result = quo;
        if rem != 0 {
            let abs_rem = rem.abs();
            let abs_div = divisor;

            let needs_inc = match rounding {
                crate::money::RoundingMode::Up => true,
                crate::money::RoundingMode::Down => false,
                crate::money::RoundingMode::HalfUp => abs_rem * 2 >= abs_div,
                crate::money::RoundingMode::HalfDown => abs_rem * 2 > abs_div,
                crate::money::RoundingMode::HalfEven => {
                    abs_rem * 2 > abs_div || (abs_rem * 2 == abs_div && quo.abs() % 2 != 0)
                }
                crate::money::RoundingMode::Ceiling => {
                    result >= 0
                }
                crate::money::RoundingMode::Floor => {
                    result < 0
                }
            };

            if needs_inc {
                result += if result >= 0 { 1 } else { -1 };
            }
        }

        Decimal {
            value: result,
            scale: decimals,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_decimal_add() {
        let a = Decimal::from_string("1.5").unwrap();
        let b = Decimal::from_string("2.3").unwrap();
        let result = a.add(&b);
        assert_eq!(result.value, 38);
        assert_eq!(result.scale, 1);
    }

    #[test]
    fn test_decimal_from_string() {
        let d = Decimal::from_string("19.99").unwrap();
        assert_eq!(d.value, 1999);
        assert_eq!(d.scale, 2);
    }
}
