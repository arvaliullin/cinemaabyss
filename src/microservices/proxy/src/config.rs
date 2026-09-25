//! Конфигурация прокси-сервиса и выбор целевого сервиса.

use std::env;

use rand::Rng;

/// Настройки proxy-сервиса и постепенной миграции.
#[derive(Clone, Debug)]
pub struct Config {
    pub port: u16,
    pub monolith_url: String,
    pub movies_service_url: String,
    pub gradual_migration: bool,
    pub movies_migration_percent: u8,
}

impl Config {
    /// Загружает конфигурацию из переменных окружения.
    pub fn from_env() -> Result<Self, String> {
        let port = env::var("PORT")
            .unwrap_or_else(|_| "8000".to_string())
            .parse::<u16>()
            .map_err(|error| format!("PORT must be a valid port: {error}"))?;

        let monolith_url = required_url("MONOLITH_URL")?;
        let movies_service_url = required_url("MOVIES_SERVICE_URL")?;

        let gradual_migration = env::var("GRADUAL_MIGRATION")
            .unwrap_or_else(|_| "false".to_string())
            .parse::<bool>()
            .map_err(|error| format!("GRADUAL_MIGRATION must be true or false: {error}"))?;

        let movies_migration_percent = env::var("MOVIES_MIGRATION_PERCENT")
            .unwrap_or_else(|_| "100".to_string())
            .parse::<u8>()
            .map_err(|error| {
                format!("MOVIES_MIGRATION_PERCENT must be an integer from 0 to 100: {error}")
            })?;

        if movies_migration_percent > 100 {
            return Err("MOVIES_MIGRATION_PERCENT must be between 0 and 100".to_string());
        }

        Ok(Self {
            port,
            monolith_url,
            movies_service_url,
            gradual_migration,
            movies_migration_percent,
        })
    }

    /// Выбирает upstream-сервис для указанного пути.
    pub fn upstream_for(&self, path: &str) -> &str {
        if !is_movies_path(path) {
            return &self.monolith_url;
        }

        if !self.gradual_migration {
            return &self.movies_service_url;
        }

        let route_to_movies = rand::rng().random_range(0..100) < self.movies_migration_percent;
        if route_to_movies {
            &self.movies_service_url
        } else {
            &self.monolith_url
        }
    }
}

/// Читает и проверяет обязательный URL из окружения.
fn required_url(name: &str) -> Result<String, String> {
    let value = env::var(name).map_err(|_| format!("{name} is required"))?;
    let value = value.trim().trim_end_matches('/');

    if value.is_empty() {
        return Err(format!("{name} must not be empty"));
    }

    reqwest::Url::parse(value).map_err(|error| format!("{name} must be a valid URL: {error}"))?;
    Ok(value.to_string())
}

/// Проверяет, относится ли путь к Movies API.
fn is_movies_path(path: &str) -> bool {
    path == "/api/movies" || path.starts_with("/api/movies/")
}
