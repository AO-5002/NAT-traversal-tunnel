package main

import (
	"NAT-tunnel-project/cmd/relay/multiplexing"
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
)

// EC2 Program

const BufferSize = 1024

var nextStreamID atomic.Uint32

func newStreamId() uint32 {
	return nextStreamID.Add(1)
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()
	fmt.Printf("Listening on %s:%s\n", protocol.DestAddr.Network, protocol.DestAddr.Port)

	// Obtain secret-key

	var secret = os.Getenv("RELAY_SECRET")
	if secret == "" {
		log.Fatal("RELAY_SECRET not set")
	}

	t := multiplexing.NewTunnelMux()
	initConn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer initConn.Close()

	// Do initial auth-handshake

	AuthHandshake(secret, initConn, t)

	// Once the initial-setup handshake is done, we can begin accepting multiple clients

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}

		// Create goroutines for every new connection & add to subscribers

		go func() {
			defer conn.Close()

			ClientSession(conn, t)

		}()
	}
}

func AuthHandshake(secret string, conn net.Conn, t *multiplexing.TunnelMux) {
	p, err := protocol.RecvPacket(conn)
	if err != nil {
		log.Fatal(err)
	}

	if multiplexing.Authorize(string(p.GetPayload()), secret) {

		// Check if client is owner & bind it to the tunnel

		if err := protocol.SendPacket(protocol.NewPacket(newStreamId(), []byte("OK")), conn); err != nil {
			log.Fatal(err)
		}

		t.BindOwner(conn)
	} else {
		if err := protocol.SendPacket(protocol.NewPacket(newStreamId(), []byte("BAD")), conn); err != nil {
			log.Fatal(err)
		}
		return
	}
}

func ClientSession(conn net.Conn, t *multiplexing.TunnelMux) {

	// Subscribe the connection to the tunnel with a newly generated id

	//c := t.Subscribe(newStreamId())

	// Read the request from the channel

}
