use anyhow::Result;
use libp2p::{
    futures::{AsyncReadExt, AsyncWriteExt, FutureExt, StreamExt},
    Multiaddr, PeerId, StreamProtocol,
};
use pnet::packet::{ipv4::Ipv4Packet, udp::UdpPacket, Packet};
use std::{collections::HashMap, str::FromStr, sync::Arc};
use tokio::{
    select,
    sync::{
        mpsc::{self, Sender},
        Mutex,
    },
};

uniffi::include_scaffolding!("lib");

pub trait PacketHandlerCallback: Send + Sync {
    fn handle_incoming_packet(&self, b: Vec<u8>);
}

#[derive(Debug, thiserror::Error)]
enum PacketHandlerError {
    #[error("Invalid packet")]
    InvalidPacket,

    #[error("Internal error: {msg}")]
    InternalError { msg: String },
}

struct PacketHandler {
    callback: Arc<dyn PacketHandlerCallback>,
    swarm: libp2p::Swarm<Behaviour>,
    streams: Mutex<HashMap<u16, Sender<Vec<u8>>>>,
}

impl PacketHandler {
    fn new(callback: Arc<dyn PacketHandlerCallback>) -> Result<Self, PacketHandlerError> {
        let stream_behaviour = libp2p_stream::Behaviour::new();
        let control = stream_behaviour.new_control();

        let swarm = libp2p::SwarmBuilder::with_new_identity()
            .with_tokio()
            .with_tcp(
                libp2p::tcp::Config::default(),
                libp2p::noise::Config::new,
                libp2p::yamux::Config::default,
            )?
            .with_relay_client(libp2p::noise::Config::new, libp2p::yamux::Config::default)?
            .with_behaviour(|_, _| Behaviour {
                // relay_client: relay_behavior,
                // ping: libp2p::ping::Behaviour::new(libp2p::ping::Config::new()),
                // identify: libp2p::identify::Behaviour::new(libp2p::identify::Config::new(
                //     "/ipfs/id/1.0.0".to_string(),
                //     key.public(),
                // )),
                http: libp2p_stream::Behaviour::new(),
            })?
            // .with_swarm_config(|c| c.with_idle_connection_timeout(Duration::from_secs(10)))
            .build();

        runtime.spawn(async {
            let result: anyhow::Result<()> = async {
                let mut swarm = libp2p::SwarmBuilder::with_new_identity()
                    .with_tokio()
                    .with_tcp(
                        libp2p::tcp::Config::default(),
                        libp2p::noise::Config::new,
                        libp2p::yamux::Config::default,
                    )?
                    .with_relay_client(libp2p::noise::Config::new, libp2p::yamux::Config::default)?
                    .with_behaviour(|_, _| Behaviour {
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

        Ok(Self {
            callback,
            control,
            runtime,
            streams: Mutex::new(HashMap::new()),
        })
    }

    async fn handle_packet(&self, b: Vec<u8>) -> Result<u8, PacketHandlerError> {
        let ip_packet = Ipv4Packet::new(&b).ok_or(PacketHandlerError::InvalidPacket)?;
        let ip_dest = ip_packet.get_destination();
        let ip_source = ip_packet.get_source();
        let ip_payload = ip_packet.payload();
        let udp_packet = UdpPacket::new(ip_payload).ok_or(PacketHandlerError::InvalidPacket)?;
        let udp_source = udp_packet.get_source();
        let udp_dest = udp_packet.get_destination();
        let payload = udp_packet.payload().to_vec();

        let mut streams = self.streams.lock().await;

        let sender = if let Some(sender) = streams.get(&udp_source) {
            sender.clone()
        } else {
            let mut control = self.control.clone();
            let (sender, mut receiver) = mpsc::channel(1024);
            streams.insert(udp_source, sender.clone());
            let callback = self.callback.clone();
            tokio::task::spawn(async move {
                let result: anyhow::Result<()> = async {
                let mut stream = control
                    .open_stream(
                        PeerId::from_str("QmShjXauvefhMYJ7tfam4adsMTy1DaxJBTKQEVh6Bs3m5i")?,
                        StreamProtocol::new("/http/1.1"),
                    )
                    .await?;
                loop {
                    let mut buf = [0; 1024];
                    select! {
                        _ = receiver.recv() => {

                            stream.write_all(&payload).await?;
                            stream.flush().await?;
                        },
                        read = stream.read(&mut buf).fuse() => {
                            match read {
                                Ok(read) => {
                                    if read == 0 {
                                        stream.close().await?;
                                        break;
                                    }
                                    let udp_packet_size = pnet::packet::udp::MutableUdpPacket::minimum_packet_size() + read;
                                    let mut udp_packet_bytes = vec![0; udp_packet_size];
                                    let mut udp_packet = pnet::packet::udp::MutableUdpPacket::new(&mut udp_packet_bytes).unwrap();
                                    udp_packet.set_destination(udp_source);
                                    udp_packet.set_source(udp_dest);
                                    udp_packet.set_payload(&buf[..read]);
                                    udp_packet.set_checksum(pnet::packet::udp::ipv4_checksum(&udp_packet.to_immutable(), &ip_source, &ip_dest));
                                    udp_packet.set_length(udp_packet_size as u16);
                                    let ip_packet_size = pnet::packet::ipv4::Ipv4Packet::minimum_packet_size() + udp_packet_size;
                                    let mut ip_packet_bytes = vec![0; ip_packet_size];
                                    let mut ip_packet = pnet::packet::ipv4::MutableIpv4Packet::new(&mut ip_packet_bytes).unwrap();
                                    ip_packet.set_version(4);
                                    ip_packet.set_header_length(5);
                                    ip_packet.set_destination(ip_source);
                                    ip_packet.set_source(ip_dest);
                                    ip_packet.set_total_length(ip_packet_size as u16);
                                    ip_packet.set_checksum(pnet::packet::ipv4::checksum(&ip_packet.to_immutable()));
                        callback.handle_incoming_packet(ip_packet.packet().to_vec());
                                    // callback(ip_packet.packet().as_ptr(), read, user_info.0);
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

        sender
            .send(udp_packet.payload().to_vec())
            .await
            .map_err(|err| PacketHandlerError::InternalError {
                msg: err.to_string(),
            });

        return Ok(1);
    }
}

#[derive(libp2p::swarm::NetworkBehaviour)]
struct Behaviour {
    // relay_client: libp2p::relay::client::Behaviour,
    // ping: libp2p::ping::Behaviour,
    // identify: libp2p::identify::Behaviour,
    http: libp2p_stream::Behaviour,
}
