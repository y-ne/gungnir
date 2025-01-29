use axum::{
    routing::{get, any},
    Router,
    Json,
    extract::{Request, ConnectInfo},
    body::{Body, to_bytes},
};
use hyper::HeaderMap;
use chrono::Utc;
use serde::Serialize;
use anyhow::Result;
use tracing::{info, Level};
use std::net::SocketAddr;

#[derive(Serialize)]
struct MessageResponse {
    message: String,
}

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    timestamp: String,
    version: String,
}

#[derive(Serialize)]
struct CallbackResponse {
    status: String,
    message: String,
    timestamp: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_target(false)
        .with_level(true)
        .with_max_level(Level::DEBUG)
        .init();

    let app = Router::new()
        .route("/", get(hello))
        .route("/health", get(health_check))
        .route("/callback", any(log_callback));

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8888").await?;
    info!("Server listening on {}", listener.local_addr()?);

    axum::serve(listener, app.into_make_service_with_connect_info::<SocketAddr>()).await?;
    Ok(())
}

async fn hello() -> Json<MessageResponse> {
    Json(MessageResponse {
        message: "hi mom".to_string(),
    })
}

async fn health_check() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        timestamp: Utc::now().to_rfc3339(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    })
}

async fn log_callback(
    method: axum::http::Method,
    headers: HeaderMap,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    req: Request<Body>,
) -> Json<CallbackResponse> {
    info!("=== Incoming Request ===");

    // Log Method
    info!("Method: {}", method);

    // Log Source Ip
    info!("Source Ip: {}", addr.ip());

    // Log raw headers
    info!("Headers: {:#?}", headers);

    // Log raw body
    if let Ok(bytes) = to_bytes(req.into_body(), usize::MAX).await {
        if let Ok(body) = String::from_utf8(bytes.to_vec()) {
            info!("Body: {}", body);
        }
    }

    info!("=== End Request ===");

    Json(CallbackResponse {
        status: "success".to_string(),
        message: "callback received".to_string(),
        timestamp: Utc::now().to_rfc3339(),
    })
}