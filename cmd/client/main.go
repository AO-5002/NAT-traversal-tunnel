package main

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
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
	defer conn.Close()

	// Retrieve URL & key (optional)
	url := os.Getenv("LOCAL_URL")
	if url == "" {
		log.Fatal(fmt.Errorf("LOCAL_URL not set"))
	}

	key := os.Getenv("RELAY_SECRET")
	authenticated := authenticate(conn, key)

	if authenticated {
		fmt.Printf("Successfully connected to: %s\n", conn.RemoteAddr())

		for {

			// Receive packet from connected-client

			p, err := protocol.RecvPacket(conn)
			if err != nil {
				fmt.Printf("Error receiving connected-client packet: %s", err)
				return
			}

			// Create a TCP connection to local-url application.

			localConn, err := net.Dial(addr.Network, url)
			if err != nil {
				fmt.Printf("Failed to connect to local-host application on URL: %s: %s\n", url, err)
				log.Fatal(err)
			}

			// Forward the request to connected local-host application

			if _, err := localConn.Write(p.GetPayload()); err != nil {
				fmt.Printf("Error forwarding payload: %s", err)
				log.Fatal(err)
			}

			// Receive whatever bytes as a response and send it back to the connected client/s

			originId := p.GetId()
			buf := make([]byte, BufferSize)
			for {
				n, err := localConn.Read(buf)
				if n > 0 {
					forwardedResponse := protocol.NewPacket(originId, buf[:n])
					protocol.SendPacket(forwardedResponse, conn)
				}
				if err != nil {
					localConn.Close()
					break
				}
			}

			// Send sentinel to signal end of response
			protocol.SendPacket(protocol.NewPacket(originId, []byte{}), conn)
			localConn.Close()
		}
	}
}

func authenticate(conn net.Conn, key string) bool {

	if key == "" {
		return false
	}

	// Create & Send packet to relay with the secret
	p := protocol.NewPacket(0, []byte(key))
	if err := protocol.SendPacket(p, conn); err != nil {
		log.Fatal(err)
	}

	// Wait for recv & parse
	p, err := protocol.RecvPacket(conn)
	if err != nil {
		return false
	}

	payload := p.GetPayload()
	if string(payload) != "OK" {
		return false
	}

	return true
}
