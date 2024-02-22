use tailscale::{TSNetwork, Tailscale};

mod graph;
mod router;

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), anyhow::Error> {
    let ts = Tailscale::new();
    ts.up().await?;
    println!("Tailscale is up!");

    let conn = rusqlite::Connection::open("data.db")?;

    // build our application with a single route
    let app = router::new_router(conn);

    ts.listen(TSNetwork::TCP, ":3000", app).await?;

    Ok(())
}
