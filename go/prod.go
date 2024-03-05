//go:build !dev

package main

import (
	"fmt"
	"net"
	"os"

	"tailscale.com/tsnet"
)

func GetListener() (net.Listener, error) {
	logfile, err := os.Create("tailscale.log")
	if err != nil {
		return nil, err
	}
	server := tsnet.Server{
		Logf: func(format string, args ...interface{}) {
			fmt.Fprintf(logfile, format, args...)
		},
	}
	ln, err := server.Listen("tcp", ":80")
	if err != nil {
		return nil, err
	}
	return ln, nil
}
