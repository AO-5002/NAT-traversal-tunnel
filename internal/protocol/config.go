package protocol

type Address struct {
	Network string
	Host    string
	Port    string
}

var DestAddr = Address{Network: "tcp", Host: "127.0.0.1", Port: "8080"}
