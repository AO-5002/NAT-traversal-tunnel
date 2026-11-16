package main

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

// EC2 Program

const BufferSize = 1024

type Hub struct {
	mu    sync.Mutex
	subs  map[net.Conn]chan []byte
	pubId net.Conn
}

func (h *Hub) subscribe(conn net.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subs[conn] = make(chan []byte, BufferSize)
}

// Send the pub's message to every sub's channel
func (h *Hub) broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subs {
		ch <- msg
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()
	fmt.Printf("Listening on %s:%s\n", protocol.DestAddr.Network, protocol.DestAddr.Port)

	secret := os.Getenv("RELAY_SECRET")
	if secret == "" {
		log.Fatal("RELAY_SECRET not set")
	}
	log.Println("secret loaded:", secret[:8]+"...")

	// Concurrency w/ Goroutines

	userCount := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		userCount++
		fmt.Printf("Accepted connection from (User #%v) %s\n", userCount, conn.RemoteAddr())

		go func() {
			defer conn.Close()

			// Read data from client
			buf := make([]byte, BufferSize)
			n, err := recv(buf, conn)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Write "ACK" back to the client
			if err := send(buf, n, conn); err != nil {
				fmt.Println(err)
			}
		}()
	}
}
