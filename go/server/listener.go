package server

import (
	"net"

	"tailscale.com/tsnet"
	"tailscale.com/types/logger"
)

func NewTcpListener() (net.Listener, error) {
	return net.ListenTCP("tcp", &net.TCPAddr{
		Port: 3000,
	})
}

func NewTailscaleListener(logf logger.Logf) (net.Listener, error) {
	server := tsnet.Server{
		Logf: logf,
	}
	ln, err := server.Listen("tcp", ":80")
	if err != nil {
		return nil, err
	}
	return ln, nil
}
