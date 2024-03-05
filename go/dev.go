//go:build dev

package main

import "net"

func GetListener() (net.Listener, error) {
	return net.ListenTCP("tcp", &net.TCPAddr{
		Port: 3000,
	})
}
