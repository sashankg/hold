package vpn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/sashankg/hold/util"
)

type Hold struct {
	host     host.Host
	conn     network.Conn
	streams  map[StreamKey]network.Stream
	callback func([]byte)
}

func Init(callback func([]byte)) *Hold {
	privKey, err := util.LoadIdentity("vpn.key")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	host, err := libp2p.New(libp2p.Identity(privKey))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	host.Network().
		Peerstore().
		AddAddr("QmYq1xnjSekRZtQnwmBtFp6CRJr4ExTr99JfVNa15ELCnR", addr, time.Hour*24)
	conn, err := host.Network().
		DialPeer(context.Background(), "QmYq1xnjSekRZtQnwmBtFp6CRJr4ExTr99JfVNa15ELCnR")
	return &Hold{
		host:     host,
		conn:     conn,
		callback: callback,
	}
}

type StreamKey struct {
	ip   string
	port layers.UDPPort
}

func (h *Hold) ProcessPacket(packetBytes []byte) error {
	packet := gopacket.NewPacket(packetBytes, layers.LayerTypeIPv4, gopacket.Default)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return errors.New("no ip layer")
	}
	ip, _ := ipLayer.(*layers.IPv4)
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return errors.New("no udp layer")
	}
	udp, _ := udpLayer.(*layers.UDP)

	streamKey := StreamKey{
		ip:   ip.SrcIP.String(),
		port: udp.SrcPort,
	}

	stream, ok := h.streams[streamKey]
	if !ok {
		stream, err := h.conn.NewStream(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
		h.streams[streamKey] = stream
		go func() {
			for {
				buf := make([]byte, 1024)
				n, err := stream.Read(buf)
				if err != nil {
					return
				}
				if n == 0 {
					delete(h.streams, streamKey)
				}
				packetBuf := gopacket.NewSerializeBuffer()

				respIp := &layers.IPv4{
					SrcIP:    ip.DstIP,
					DstIP:    ip.SrcIP,
					Protocol: layers.IPProtocolUDP,
					Version:  4,
				}
				respUdp := &layers.UDP{
					SrcPort: udp.DstPort,
					DstPort: udp.SrcPort,
				}
				respUdp.SetNetworkLayerForChecksum(respIp)
				err = gopacket.SerializeLayers(
					packetBuf,
					gopacket.SerializeOptions{
						ComputeChecksums: true,
						FixLengths:       true,
					},
					respIp,
					respUdp,
					gopacket.Payload(buf),
				)
				h.callback(packetBuf.Bytes())
			}
		}()
	}
	_, err := stream.Write(udp.Payload)
	return err
}
