use axum::routing::*;

pub fn new_router() -> axum::Router {
    Router::new().route("/", get(|| async { "Hello, World!" }))
}
