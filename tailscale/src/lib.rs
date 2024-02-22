#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]

use std::{
    ffi::{c_int, CStr, CString},
    fs::OpenOptions,
    os::fd::{AsRawFd, FromRawFd},
};

use hyper::server::conn::http2::Builder;
use hyper_util::{
    rt::{TokioExecutor, TokioIo},
    service::TowerToHyperService,
};
use tokio::task::spawn_blocking;

include!(concat!(env!("OUT_DIR"), "/libtailscale.rs"));

pub enum TSNetwork {
    TCP,
    UDP,
}

impl ToString for TSNetwork {
    fn to_string(&self) -> String {
        match self {
            Self::TCP => "tcp".to_string(),
            Self::UDP => "udp".to_string(),
        }
    }
}

#[derive(Clone)]
pub struct Tailscale {
    ts: tailscale,
}

impl Tailscale {
    pub fn new() -> Self {
        unsafe {
            let ts = tailscale_new();
            let file = OpenOptions::new()
                .write(true)
                .create(true)
                .open("tailscale.log")
                .expect("failed to open log file");
            let fd = file.as_raw_fd();
            tailscale_set_logfd(ts, fd);
            Tailscale { ts }
        }
    }

    pub fn new_with_args(
        dir: &str,
        hostname: &str,
        auth_key: &str,
        control_url: &str,
        ephemeral: bool,
    ) -> Self {
        let dir = CString::new(dir).unwrap();
        let hostname = CString::new(hostname).unwrap();
        let auth_key = CString::new(auth_key).unwrap();
        let control_url = CString::new(control_url).unwrap();
        unsafe {
            let ts = tailscale_new();
            tailscale_set_dir(ts, dir.as_ptr());
            tailscale_set_hostname(ts, hostname.as_ptr());
            tailscale_set_authkey(ts, auth_key.as_ptr());
            tailscale_set_control_url(ts, control_url.as_ptr());
            tailscale_set_ephemeral(ts, ephemeral.into());
            Tailscale { ts }
        }
    }

    pub async fn up(&self) -> anyhow::Result<()> {
        let ts = self.ts;
        tokio::task::spawn_blocking(move || handle_error(ts, unsafe { tailscale_up(ts) })).await?
    }

    pub async fn listen(
        &self,
        network: TSNetwork,
        addr: &str,
        service: axum::Router,
    ) -> anyhow::Result<()> {
        let network = CString::new(network.to_string()).unwrap();
        let addr = CString::new(addr).unwrap();
        let mut listener = 0;
        handle_error(self.ts, unsafe {
            tailscale_listen(self.ts, network.as_ptr(), addr.as_ptr(), &mut listener)
        })?;
        println!("{:?}", listener);

        loop {
            let conn_out = {
                let ts = self.ts;
                spawn_blocking(move || {
                    let mut conn_out = 0;
                    handle_error(ts, unsafe { tailscale_accept(listener, &mut conn_out) })?;
                    Ok::<i32, anyhow::Error>(conn_out)
                })
                .await?
            }?;

            let tcp_stream = unsafe { std::net::TcpStream::from_raw_fd(conn_out) };
            tcp_stream.set_nonblocking(true)?;
            let tcp_stream = TokioIo::new(tokio::net::TcpStream::from_std(tcp_stream)?);

            let service = service.clone();
            tokio::task::spawn(async move {
                Builder::new(TokioExecutor::new())
                    .serve_connection(tcp_stream, TowerToHyperService::new(service))
                    .await
            });
        }
    }
}

fn handle_error(ts: tailscale, code: c_int) -> anyhow::Result<()> {
    if code != -1 {
        return Ok(());
    }
    let mut buffer: [i8; 256] = [0; 256];
    let ret = unsafe { tailscale_errmsg(ts, &mut buffer as *mut _, buffer.len()) };
    if ret != 0 {
        return Err(anyhow::anyhow!(
            "tailscale internal error: failed to get error message"
        ));
    }
    println!("{:?}", buffer);
    let cstr = unsafe { CStr::from_ptr(buffer.as_ptr()) };
    Err(anyhow::anyhow!(cstr.to_string_lossy()))
}

impl Default for Tailscale {
    fn default() -> Self {
        Self::new()
    }
}

impl Drop for Tailscale {
    fn drop(&mut self) {
        unsafe {
            tailscale_close(self.ts);
        }
    }
}
