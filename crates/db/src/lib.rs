use std::sync::Arc;

use sqlx::{
    PgTransaction, Postgres,
    error::{DatabaseError, ErrorKind},
    postgres::PgPoolOptions,
};
use uwececa_macros::transitive_from;

pub mod testing;

pub type Result<T> = std::result::Result<T, Error>;

pub type Db = Arc<sqlx::PgPool>;
pub type Tx<'a> = PgTransaction<'a>;
pub trait Executor<'a>: sqlx::Executor<'a, Database = Postgres> {}

#[derive(Debug)]
pub enum Error {
    UniqueViolation,
    ForeignKeyViolation,
    NotNullViolation,
    CheckViolation,
    RowNotFound,
    MigrateError,
    Unknown(color_eyre::Report),
}

impl From<Box<dyn DatabaseError + 'static>> for Error {
    fn from(value: Box<dyn DatabaseError>) -> Self {
        match value.kind() {
            ErrorKind::UniqueViolation => Self::UniqueViolation,
            ErrorKind::ForeignKeyViolation => Self::ForeignKeyViolation,
            ErrorKind::NotNullViolation => Self::NotNullViolation,
            ErrorKind::CheckViolation => Self::CheckViolation,
            _ => Self::Unknown(value.into()),
        }
    }
}

impl From<sqlx::Error> for Error {
    fn from(value: sqlx::Error) -> Self {
        match value {
            sqlx::Error::Database(e) => e.into(),
            sqlx::Error::RowNotFound => Self::RowNotFound,
            _ => Self::Unknown(value.into()),
        }
    }
}

impl From<color_eyre::Report> for Error {
    fn from(value: color_eyre::Report) -> Self {
        Self::Unknown(value)
    }
}

transitive_from!(sqlx::migrate::MigrateError, Error, color_eyre::Report);

pub async fn connect(url: &str) -> Result<Db> {
    let pool = PgPoolOptions::new().connect(url).await?;

    sqlx::migrate!("../../migrations").run(&pool).await?;

    Ok(Arc::new(pool))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_conn() {
        let _ = testing::get_test_tx();
    }
}
