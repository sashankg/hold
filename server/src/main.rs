use std::sync::Arc;

use hyper_util::{rt::TokioIo, service::TowerToHyperService};

mod graph;
mod router;

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), anyhow::Error> {
    let conn = rusqlite::Connection::open("data.db")?;

    // build our application with a single route
    let app = router::new_router(conn);

    start_server(app).await
}

#[cfg(feature = "prod")]
async fn start_server(app: axum::Router) -> anyhow::Result<()> {
    use tailscale::{TSNetwork, Tailscale};
    let ts = Arc::new(Tailscale::new());
    ts.up().await?;
    println!("Tailscale is up!");
    let listener = ts.listen(TSNetwork::TCP, ":3000")?;
    loop {
        let conn = ts.accept(listener).await?;

        let service = app.clone();
        tokio::task::spawn(async move {
            // hyper::server::conn::http2::Builder::new(TokioExecutor::new())
            hyper::server::conn::http1::Builder::new()
                .serve_connection(
                    TokioIo::new(tokio::net::TcpStream::from_std(conn).unwrap()),
                    TowerToHyperService::new(service),
                )
                .await
        });
    }
}

#[cfg(feature = "dev")]
async fn start_server(app: axum::Router) -> anyhow::Result<()> {
    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await?;
    axum::serve(listener, app).await;
    Ok(())
}
