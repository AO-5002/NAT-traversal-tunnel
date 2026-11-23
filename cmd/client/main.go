package main

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

const BufferSize = 64 * 1024

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

		var mu sync.Mutex

		for {

			// Receive packet from connected-client

			p, err := protocol.RecvPacket(conn)
			if err != nil {
				fmt.Printf("Error receiving connected-client packet: %s", err)
				return
			}

			// Spawn a goroutine per request so long-lived connections don't block the loop

			go func(p *protocol.Packet) {
				originId := p.GetId()

				localConn, err := net.Dial(addr.Network, url)
				if err != nil {
					fmt.Printf("Failed to connect to local-host application on URL: %s: %s\n", url, err)
					return
				}
				defer localConn.Close()

				if _, err := localConn.Write(p.GetPayload()); err != nil {
					fmt.Printf("Error forwarding payload: %s", err)
					return
				}

				buf := make([]byte, BufferSize)
				for {
					n, err := localConn.Read(buf)
					if n > 0 {
						mu.Lock()
						protocol.SendPacket(protocol.NewPacket(originId, buf[:n]), conn)
						mu.Unlock()
					}
					if err != nil {
						break
					}
				}

				// Send sentinel to signal end of response
				mu.Lock()
				protocol.SendPacket(protocol.NewPacket(originId, []byte{}), conn)
				mu.Unlock()
			}(p)
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
