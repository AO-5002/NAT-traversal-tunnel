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

	// Create initial stream with sender

	t := multiplexing.NewTunnelMux()
	initConn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer initConn.Close()

	// Do initial auth-handshake

	if !AuthHandshake(secret, initConn) {
		log.Fatal("Auth handshake failed")
	}

	// Bind the sender-stream as the owner of the tunnel

	t.BindOwner(initConn)
	go t.Demux()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}

		// Create goroutines for every new connection & add to subscribers

		go func(conn net.Conn) {
			defer conn.Close()
			ClientSession(conn, t)
		}(conn)
	}
}

func AuthHandshake(secret string, conn net.Conn) bool {
	p, err := protocol.RecvPacket(conn)
	if err != nil {
		log.Fatal(err)
	}

	if multiplexing.Authorize(string(p.GetPayload()), secret) {
		if err := protocol.SendPacket(protocol.NewPacket(newStreamId(), []byte("OK")), conn); err != nil {
			log.Fatal(err)
		}
		return true
	}

	if err := protocol.SendPacket(protocol.NewPacket(newStreamId(), []byte("BAD")), conn); err != nil {
		log.Fatal(err)
	}
	return false
}

func ClientSession(conn net.Conn, t *multiplexing.TunnelMux) {

	// Subscribe the connection to the tunnel with a newly generated id

	id := newStreamId()
	c := t.Subscribe(id)

	// Receive request from connected client/s

	buf := make([]byte, BufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Forward raw bytes to owner-connection

		t.Mux(id, buf[:n])

		// Receive responses from owner-connection
		response := <-c
		if _, err := conn.Write(response); err != nil {
			log.Fatal(err)
		}
	}
}
