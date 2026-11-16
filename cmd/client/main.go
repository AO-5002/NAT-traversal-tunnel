package main

import (
	"NAT-tunnel-project/internal/protocol"
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
)

const BufferSize = 1024

func main() {

	if err := godotenv.Load("internal/protocol/.env"); err != nil {
		log.Fatal(err)
	}

	// Get the .env properties from the local .env file in /internal/protocol
	addr := protocol.Address{
		Network: os.Getenv("DEST_IPV4_NETWORK"),
		Host:    os.Getenv("DEST_IPV4_HOST"),
		Port:    os.Getenv("DEST_IPV4_PORT"),
	}

	// Connect to the relay-server in the EC2 instance.
	conn, err := net.Dial(addr.Network, net.JoinHostPort(addr.Host, addr.Port))
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
