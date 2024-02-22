use std::sync::Arc;

use hyper_util::{
    rt::{TokioExecutor, TokioIo},
    service::TowerToHyperService,
};
use tailscale::{TSNetwork, Tailscale};

mod graph;
mod router;

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), anyhow::Error> {
    let ts = Arc::new(Tailscale::new());
    ts.up().await?;
    println!("Tailscale is up!");

    let conn = rusqlite::Connection::open("data.db")?;

    // build our application with a single route
    let app = router::new_router(conn);

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
