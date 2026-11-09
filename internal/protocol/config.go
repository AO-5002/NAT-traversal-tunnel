package protocol

import (
	"os"

	"github.com/joho/godotenv"
)

var DestAddr Address

type Address struct {
	Network string
	Host    string
	Port    string
}

// Configure & Initialize .env variables into the destination-address struct.
func init() {
	_ = godotenv.Load()
	envNetwork := os.Getenv("DEST_IPV4_NETWORK")
	envHost := os.Getenv("DEST_IPV4_HOST")
	envPort := os.Getenv("DEST_IPV4_PORT")
	DestAddr = Address{Network: envNetwork, Host: envHost, Port: envPort}
}
