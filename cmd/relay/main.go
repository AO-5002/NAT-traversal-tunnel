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
	ln, err := net.Listen(protocol.DestAddr.Network, ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("Listening on %s:%s\n", protocol.DestAddr.Network, protocol.DestAddr.Port)

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	buf := make([]byte, BufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Connection closed.")
		} else {
			fmt.Println(err)
		}
		return
	}

	relayMessage := fmt.Sprintf("ACK %s", buf[:n])
	_, err = conn.Write([]byte(relayMessage))
	if err != nil {
		fmt.Println(err)
		return
	}
}
