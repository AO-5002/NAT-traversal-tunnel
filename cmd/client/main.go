package main

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

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
		_, err := http.Post(url, "application/json", nil)
		if err != nil {
			fmt.Println(err)
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
