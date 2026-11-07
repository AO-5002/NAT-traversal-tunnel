package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

const BufferSize int = 1024

type Address struct {
	network string
	address string
	port    string
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, BufferSize)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected.")
			} else {
				log.Println(err)
			}
			return
		}

		relayMessage := fmt.Sprintf("Server echoes: %s", buf[:n])
		_, err = conn.Write([]byte(relayMessage))
		if err != nil {
			log.Println(err)
			return
		}

	}
}

func main() {

	// Create address we want to listen on
	addr := Address{network: "tcp", address: "127.0.0.1", port: "8080"}

	// Create, bind, listen on this address
	ln, err := net.Listen(addr.network, net.JoinHostPort(addr.address, addr.port))
	defer ln.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening on " + addr.address + ":" + addr.port + "...")

	// Accept incoming connection

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		handleConnection(conn)
	}
}
