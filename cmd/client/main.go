package main

import (
	"NAT-tunnel-project/internal/protocol"
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"

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
	localURL := os.Getenv("LOCAL_URL")
	if localURL == "" {
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

				// Parse the raw HTTP request bytes into a proper request
				req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(p.GetPayload())))
				if err != nil {
					fmt.Printf("Error parsing request: %s\n", err)
					return
				}

				// Retarget to local service with correct host
				req.RequestURI = ""
				req.URL.Host = localURL
				req.URL.Scheme = "http"

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Error forwarding request: %s\n", err)
					return
				}
				defer resp.Body.Close()

				var buf bytes.Buffer
				resp.Write(&buf)
				data := buf.Bytes()

				// Send response in chunks
				for len(data) > 0 {
					chunk := data
					if len(chunk) > BufferSize {
						chunk = data[:BufferSize]
					}
					mu.Lock()
					protocol.SendPacket(protocol.NewPacket(originId, chunk), conn)
					mu.Unlock()
					data = data[len(chunk):]
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
