package main

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"io"
	"log"
	"net"
)

// EC2 Program

const BufferSize = 1024

func main() {
	ln, err := net.Listen(protocol.DestAddr.Network, ":"+protocol.DestAddr.Port)
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()
	fmt.Printf("Listening on %s:%s\n", protocol.DestAddr.Network, protocol.DestAddr.Port)

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

func recv(buf []byte, conn net.Conn) (int, error) {
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Connection closed.")
		}
		return 0, err
	}

	return n, nil
}

func send(buf []byte, n int, conn net.Conn) error {
	relayMessage := fmt.Sprintf("ACK %s", buf[:n])
	_, err := conn.Write([]byte(relayMessage))
	return err
}
