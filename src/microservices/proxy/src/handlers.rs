//! HTTP-обработчики прокси-сервиса.

use axum::Json;
use axum::body::{Body, to_bytes};
use axum::extract::{Request, State};
use axum::http::{HeaderMap, HeaderName, StatusCode};
use axum::response::{IntoResponse, Response};

use crate::AppState;

const MAX_REQUEST_BODY_SIZE: usize = 16 * 1024 * 1024;

/// Возвращает состояние работоспособности proxy-сервиса.
pub async fn health() -> Json<serde_json::Value> {
    Json(serde_json::json!({ "status": true }))
}

/// Перенаправляет входящий HTTP-запрос в выбранный upstream-сервис.
pub async fn proxy(State(state): State<AppState>, request: Request) -> Response {
    let (parts, body) = request.into_parts();
    let path_and_query = parts
        .uri
        .path_and_query()
        .map(|value| value.as_str())
        .unwrap_or("/");
    let upstream = state.config.upstream_for(parts.uri.path());
    let target_url = format!("{upstream}{path_and_query}");

    let body = match to_bytes(body, MAX_REQUEST_BODY_SIZE).await {
        Ok(body) => body,
        Err(error) => {
            return (
                StatusCode::PAYLOAD_TOO_LARGE,
                format!("failed to read request body: {error}"),
            )
                .into_response();
        }
    };

    println!(
        "proxying {} {} to {}",
        parts.method, path_and_query, upstream
    );

    let upstream_response = match state
        .client
        .request(parts.method, target_url)
        .headers(forwarded_headers(&parts.headers, true))
        .body(body)
        .send()
        .await
    {
        Ok(response) => response,
        Err(error) => {
            return (
                StatusCode::BAD_GATEWAY,
                format!("upstream request failed: {error}"),
            )
                .into_response();
        }
    };

    let status = upstream_response.status();
    let headers = forwarded_headers(upstream_response.headers(), false);
    let body = match upstream_response.bytes().await {
        Ok(body) => body,
        Err(error) => {
            return (
                StatusCode::BAD_GATEWAY,
                format!("failed to read upstream response: {error}"),
            )
                .into_response();
        }
    };

    let mut response = Response::new(Body::from(body));
    *response.status_mut() = status;
    *response.headers_mut() = headers;
    response
}

/// Копирует заголовки, исключая служебные заголовки соединения.
fn forwarded_headers(source: &HeaderMap, skip_host: bool) -> HeaderMap {
    let mut target = HeaderMap::new();

    for (name, value) in source {
        if (skip_host && name == axum::http::header::HOST) || is_hop_by_hop(name) {
            continue;
        }
        target.append(name.clone(), value.clone());
    }

    target
}

/// Проверяет, относится ли заголовок к текущему HTTP-соединению.
fn is_hop_by_hop(name: &HeaderName) -> bool {
    matches!(
        name.as_str(),
        "connection"
            | "keep-alive"
            | "proxy-authenticate"
            | "proxy-authorization"
            | "proxy-connection"
            | "te"
            | "trailer"
            | "transfer-encoding"
            | "upgrade"
    )
}
