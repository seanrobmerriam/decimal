use wasm_bindgen::prelude::*;

#[derive(Debug, Clone)]
pub enum ErrorKind {
    CurrencyMismatch,
    DivisionByZero,
    Overflow,
    Underflow,
    PrecisionLoss,
    ParseError,
    InvalidOperation,
    Rounding,
}

#[wasm_bindgen]
#[derive(Debug, Clone)]
pub struct Error {
    kind: ErrorKind,
    message: String,
}

#[wasm_bindgen]
impl Error {
    #[wasm_bindgen(constructor)]
    pub fn new(message: &str) -> Error {
        Error {
            kind: ErrorKind::InvalidOperation,
            message: message.to_string(),
        }
    }

    #[wasm_bindgen(getter)]
    pub fn kind(&self) -> String {
        format!("{:?}", self.kind)
    }

    #[wasm_bindgen(getter)]
    pub fn message(&self) -> String {
        self.message.clone()
    }
}

impl Error {
    pub fn new_currency_mismatch(left: &str, right: &str, operation: &str) -> Error {
        Error {
            kind: ErrorKind::CurrencyMismatch,
            message: format!(
                "currency mismatch: cannot {} {} and {}",
                operation, left, right
            ),
        }
    }

    pub fn new_division_by_zero() -> Error {
        Error {
            kind: ErrorKind::DivisionByZero,
            message: "division by zero".to_string(),
        }
    }

    pub fn new_overflow(operation: &str) -> Error {
        Error {
            kind: ErrorKind::Overflow,
            message: format!("overflow in operation: {}", operation),
        }
    }

    pub fn new_parse_error(input: &str, reason: &str) -> Error {
        Error {
            kind: ErrorKind::ParseError,
            message: format!("parse error: {} - input: {}", reason, input),
        }
    }

    pub fn new_invalid_operation(reason: &str) -> Error {
        Error {
            kind: ErrorKind::InvalidOperation,
            message: reason.to_string(),
        }
    }
}
