package main

import (
	"fmt"
	"io"
	"net"
)

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
