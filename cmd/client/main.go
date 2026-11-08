package main

import (
	"NAT-tunnel-project/internal/protocol"
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const BufferSize = 1024

func main() {

	conn, err := net.Dial(protocol.DestAddr.Network, net.JoinHostPort(protocol.DestAddr.Host, protocol.DestAddr.Port))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to: %s\n", conn.RemoteAddr())
	fmt.Print("Write something: ")

	defer conn.Close()
	buf := make([]byte, BufferSize)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println(err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
			}
			fmt.Println(err)
		}

		fmt.Println(string(buf[:n]))
		break
	}

}
