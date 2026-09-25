//! Точка входа proxy.

mod config;
mod handlers;

use axum::Router;
use axum::routing::get;

use config::Config;

#[derive(Clone)]
struct AppState {
    config: Config,
    client: reqwest::Client,
}

#[tokio::main]
async fn main() {
    let config =
        Config::from_env().unwrap_or_else(|error| panic!("invalid proxy configuration: {error}"));
    let address = format!("0.0.0.0:{}", config.port);

    let client = reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .build()
        .expect("failed to create HTTP client");

    let app = Router::new()
        .route("/health", get(handlers::health))
        .fallback(handlers::proxy)
        .with_state(AppState { config, client });

    let listener = tokio::net::TcpListener::bind(&address)
        .await
        .unwrap_or_else(|error| panic!("failed to bind {address}: {error}"));

    println!("proxy service listening on {address}");
    axum::serve(listener, app)
        .await
        .expect("proxy server stopped unexpectedly");
}
