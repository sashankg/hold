package main

import (
	"database/sql"
	"embed"
	"fmt"

	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/sashankg/hold/util"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	goose.SetBaseFS(migrations)

	_, err := sql.Open("sqlite3", "rendezvous.db")

	privKey, err := util.LoadIdentity("rendezvous.key")
	if err != nil {
		panic(err)
	}
	host, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(websocket.New),
		),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",
			"/ip4/0.0.0.0/tcp/4002/ws",
			// "/ip4/0.0.0.0/tcp/4001/wss",
		),
		libp2p.ForceReachabilityPublic(),
		libp2p.EnableRelayService(relay.WithInfiniteLimits()),
	)
	if err != nil {
		panic(err)
	}

	println("Host ID", host.ID().String())

	for _, addr := range host.Network().ListenAddresses() {
		println("Listening on", addr.String())
	}
	fmt.Println("%w", logging.GetSubsystems())

	util.WaitForInterrupt()

	host.Close()
}

type DbPeerStore struct{}

// var _ peerstore.Peerstore = (*DbPeerStore)(nil)
