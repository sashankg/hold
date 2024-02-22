use axum::*;

pub async fn handler(_: body::Bytes) -> Result<String, anyhow::Error> {
    Ok("hello world".to_string())
}
