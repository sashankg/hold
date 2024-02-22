#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]

use std::{
    ffi::{c_int, CStr, CString},
    fs::OpenOptions,
    net::TcpStream,
    os::fd::{AsRawFd, FromRawFd},
    sync::Arc,
};

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

#[derive(Clone, Copy)]
pub struct TailscaleListener {
    inner: i32,
}

pub struct Tailscale {
    inner: tailscale,
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
            Tailscale { inner: ts }
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
            Tailscale { inner: ts }
        }
    }

    pub async fn up(self: &Arc<Self>) -> anyhow::Result<()> {
        let ts = self.clone();
        tokio::task::spawn_blocking(move || ts.handle_error(unsafe { tailscale_up(ts.inner) }))
            .await?
    }

    pub fn listen(&self, network: TSNetwork, addr: &str) -> anyhow::Result<TailscaleListener> {
        let network = CString::new(network.to_string()).unwrap();
        let addr = CString::new(addr).unwrap();
        let mut listener = 0;
        self.handle_error(unsafe {
            tailscale_listen(self.inner, network.as_ptr(), addr.as_ptr(), &mut listener)
        })?;
        println!("Listening on {:?}", listener);
        Ok(TailscaleListener { inner: listener })
    }

    pub async fn accept(
        self: &Arc<Self>,
        listener: TailscaleListener,
    ) -> anyhow::Result<TcpStream> {
        let ts = self.clone();
        let conn = tokio::task::spawn_blocking(move || {
            let mut conn_out = 0;
            ts.handle_error(unsafe { tailscale_accept(listener.inner, &mut conn_out) })?;
            anyhow::Ok(conn_out)
        })
        .await??;
        let tcp_stream = unsafe { std::net::TcpStream::from_raw_fd(conn) };
        tcp_stream.set_nonblocking(true)?;
        Ok(tcp_stream)
    }

    fn handle_error(&self, code: c_int) -> anyhow::Result<()> {
        if code != -1 {
            return Ok(());
        }
        let mut buffer: [i8; 256] = [0; 256];
        let ret = unsafe { tailscale_errmsg(self.inner, &mut buffer as *mut _, buffer.len()) };
        if ret != 0 {
            return Err(anyhow::anyhow!(
                "tailscale internal error: failed to get error message"
            ));
        }
        println!("{:?}", buffer);
        let cstr = unsafe { CStr::from_ptr(buffer.as_ptr()) };
        Err(anyhow::anyhow!(cstr.to_string_lossy()))
    }
}

impl Default for Tailscale {
    fn default() -> Self {
        Self::new()
    }
}

impl Drop for Tailscale {
    fn drop(&mut self) {
        unsafe {
            tailscale_close(self.inner);
        }
    }
}
