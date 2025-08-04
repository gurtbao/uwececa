use std::sync::{Arc, LazyLock};

use serde::Deserialize;

pub type Config = Arc<Cfg>;

#[derive(Debug, Deserialize)]
pub struct Cfg {
    pub web_port: u16,
    pub assets_location: String,
    pub template_path: String,
}

const GIT_COMMIT_HASH: &'static str = env!("GIT_HASH");

pub const fn git_hash() -> &'static str {
    GIT_COMMIT_HASH
}

#[derive(Debug, Clone, Copy)]
#[repr(u8)]
pub enum Environment {
    Development,
    Production,
}

impl Into<&'static str> for Environment {
    fn into(self) -> &'static str {
        match self {
            Self::Development => "development",
            Self::Production => "production",
        }
    }
}

static ENVIRONMENT: LazyLock<Environment> = LazyLock::new(|| {
    let is_development = matches!(
        std::env::var("UWECECA_DEVELOPMENT").as_deref(),
        Ok("1") | Ok("true")
    );

    if is_development {
        Environment::Development
    } else {
        Environment::Production
    }
});

#[inline]
pub fn is_production() -> bool {
    matches!(*ENVIRONMENT, Environment::Production)
}

#[inline]
pub fn is_development() -> bool {
    matches!(*ENVIRONMENT, Environment::Development)
}

#[inline]
pub fn get_environment() -> Environment {
    *ENVIRONMENT
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_git_hash() {
        assert_ne!(git_hash().len(), 0);
    }
}
