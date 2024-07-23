use std::{ffi::c_uint, os::raw::c_uchar, str::FromStr};

use futures::{channel::mpsc, select, AsyncReadExt, AsyncWriteExt, FutureExt, SinkExt, StreamExt};
use libp2p::{Multiaddr, PeerId, StreamProtocol};
use pnet::packet::{MutablePacket, Packet};
use tracing::Level;

#[no_mangle]
pub unsafe extern "C" fn handle_packets(hold: *mut Hold, bytes: *mut c_uchar, len: c_uint) -> i32 {
    let hold = unsafe { &mut *hold };
    let bytes = unsafe { std::slice::from_raw_parts(bytes as *const u8, len as usize) };
    let rt = &hold.runtime;
    match rt.block_on(handle_packet(&mut hold.context, bytes)) {
        Ok(_) => 0,
        Err(err) => {
            println!("Error: {:?}", err);
            1
        }
    }
}

async fn handle_packet(hold: &mut Context, bytes: &[u8]) -> Result<(), anyhow::Error> {
    println!("handle_packet");
    let ip_packet =
        pnet::packet::ipv4::Ipv4Packet::new(bytes).ok_or(anyhow::anyhow!("Invalid ip packet"))?;
    let ip_dest = ip_packet.get_destination();
    let ip_source = ip_packet.get_source();
    let ip_payload = ip_packet.payload();
    let udp_packet = pnet::packet::udp::UdpPacket::new(ip_payload)
        .ok_or(anyhow::anyhow!("Invalid udp packet"))?;
    let udp_source = udp_packet.get_source();
    let udp_dest = udp_packet.get_destination();
    let payload = udp_packet.payload().to_vec();

    let mut sender = if let Some(sender) = hold.streams.get(&udp_source) {
        sender.clone()
    } else {
        let (sender, mut receiver) = mpsc::channel(1024);
        hold.streams.insert(udp_source, sender.clone());
        let callback = hold.callback;
        // copy bytes to a new buffer
        let mut bytes = bytes.to_vec();
        let mut control = hold.control.clone();
        tokio::task::spawn(async move {
            let result: anyhow::Result<()> = async {
                println!("opening stream");
                let mut stream = control
                    .open_stream(
                        PeerId::from_str("QmShjXauvefhMYJ7tfam4adsMTy1DaxJBTKQEVh6Bs3m5i")?,
                        StreamProtocol::new("/http/1.1"),
                    )
                    .await? ;
                println!("opened stream");
                loop {
                    let mut buf = [0; 1024];
                    println!("reading");
                    select! {
                        _ = receiver.next() => {

                            stream.write_all(&payload).await?;
                            stream.flush().await?;
                        },
                        read = stream.read(&mut buf).fuse() => {
                            match read {
                                Ok(read) => {
                                    println!("read: {:?}", buf);
                                    let mut resp_packet = pnet::packet::ipv4::MutableIpv4Packet::new(&mut bytes).unwrap();
                                    let  ip_payload = resp_packet.payload_mut();
                                    let mut udp_packet = pnet::packet::udp::MutableUdpPacket::new(ip_payload).unwrap();
                                    udp_packet.set_destination(udp_source);
                                    udp_packet.set_source(udp_dest);
                                    udp_packet.set_payload(&buf);
                                    udp_packet.set_checksum(pnet::packet::udp::ipv4_checksum(&udp_packet.to_immutable(), &ip_source, &ip_dest));
                                    resp_packet.set_destination(ip_source);
                                    resp_packet.set_source(ip_dest);
                                    // / resp_packet.set_tota&l_length(udp_packet.packet().len() as u16);
                                    resp_packet.set_checksum(pnet::packet::ipv4::checksum(&resp_packet.to_immutable()));
                                    callback(buf.as_mut_ptr(), read.try_into()?);
                                },
                                Err(err) => {
                                    println!("error: {:?}", err);
                                    break;
                                }
                            }
                        }
                    }
                }
                Ok(())
            }.await;
            println!("result: {:?}", result);
        });
        sender
    };
    sender.send(udp_packet.payload().to_vec()).await?;
    Ok(())
}

#[no_mangle]
pub extern "C" fn init(callback: extern "C" fn(*mut c_uchar, c_uint)) -> *mut Hold {
    println!("init");
    tracing_subscriber::fmt::init();
    let rt = tokio::runtime::Runtime::new().unwrap();
    let control = init_libp2p(&rt).unwrap();
    Box::into_raw(Box::new(Hold {
        context: Context {
            control,
            streams: std::collections::HashMap::new(),
            callback,
        },
        runtime: rt,
    }))
}

struct Context {
    control: libp2p_stream::Control,
    streams: std::collections::HashMap<u16, mpsc::Sender<Vec<u8>>>,
    callback: extern "C" fn(*mut c_uchar, c_uint),
}

pub struct Hold {
    context: Context,
    runtime: tokio::runtime::Runtime,
}

#[derive(libp2p::swarm::NetworkBehaviour)]
struct Behaviour {
    // relay_client: libp2p::relay::client::Behaviour,
    // ping: libp2p::ping::Behaviour,
    // identify: libp2p::identify::Behaviour,
    http: libp2p_stream::Behaviour,
}

#[tracing::instrument]
fn init_libp2p(rt: &tokio::runtime::Runtime) -> anyhow::Result<libp2p_stream::Control> {
    tracing::event!(Level::INFO, "inside my_function!");
    let stream_behaviour = libp2p_stream::Behaviour::new();
    let control = stream_behaviour.new_control();

    rt.spawn(async {
        let result: anyhow::Result<()> = async {
            let mut swarm = libp2p::SwarmBuilder::with_new_identity()
                .with_tokio()
                .with_tcp(
                    libp2p::tcp::Config::default(),
                    libp2p::noise::Config::new,
                    libp2p::yamux::Config::default,
                )?
                .with_relay_client(libp2p::noise::Config::new, libp2p::yamux::Config::default)?
                .with_behaviour(|key, relay_behavior| Behaviour {
                    // relay_client: relay_behavior,
                    // ping: libp2p::ping::Behaviour::new(libp2p::ping::Config::new()),
                    // identify: libp2p::identify::Behaviour::new(libp2p::identify::Config::new(
                    //     "/ipfs/id/1.0.0".to_string(),
                    //     key.public(),
                    // )),
                    http: stream_behaviour,
                })?
                // .with_swarm_config(|c| c.with_idle_connection_timeout(Duration::from_secs(10)))
                .build();

            swarm.dial(Multiaddr::from_str(
                "/ip4/127.0.0.1/tcp/4001/p2p/QmShjXauvefhMYJ7tfam4adsMTy1DaxJBTKQEVh6Bs3m5i",
            )?)?;
            println!("Listening on: {:?}", swarm.connected_peers().next());
            println!("network info: {:?}", swarm.network_info());

            loop {
                let event = swarm.next().await;
                println!("event: {:?}", event);
            }
        }
        .await;
        println!("event loop result: {:?}", result);
    });

    Ok(control)
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_handle_packets() {
        tracing_subscriber::fmt::init();
    }

    extern "C" fn test_callback(data: *mut c_uchar, len: c_uint) {
        println!("callback");
    }
}
