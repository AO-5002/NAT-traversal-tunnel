package protocol

var DestAddr Address

type Address struct {
	Network string
	Host    string
	Port    string
}
