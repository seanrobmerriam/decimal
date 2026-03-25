mod decimal;
mod errors;
mod money;

pub use decimal::Decimal;
pub use errors::{Error, ErrorKind};
pub use money::{Currency, Money, RoundingMode};

pub mod prelude {
    pub use crate::decimal::Decimal;
    pub use crate::errors::{Error, ErrorKind};
    pub use crate::money::{Currency, Money, RoundingMode};
}
