package multiplexing

import (
	"NAT-tunnel-project/internal/protocol"
	"fmt"
	"net"
	"sync"
)

const BufferSize = 1024

type TunnelMux struct {
	mu       sync.Mutex
	subs     map[uint32]chan []byte
	ownerCon net.Conn
}

func NewTunnelMux() *TunnelMux {
	return &TunnelMux{
		subs: make(map[uint32]chan []byte),
	}
}

// Subscribe Client connections that are not the publisher subscribe to the hub, and get their channel.
func (h *TunnelMux) Subscribe(streamId uint32) <-chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	c := make(chan []byte, BufferSize)
	h.subs[streamId] = c
	return c
}

// BindOwner Bind the original connection to the hub.
func (h *TunnelMux) BindOwner(con net.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ownerCon = con
}

// Mux (Clients -> Owner)
func (h *TunnelMux) Mux(id uint32, payload []byte) {

	// Multiplexing action

	p := protocol.NewPacket(id, payload)
	h.mu.Lock()
	err := protocol.SendPacket(p, h.ownerCon)
	if err != nil {
		fmt.Println(err)
	}
	h.mu.Unlock()
}

// Demux (Owner -> Clients)
func (h *TunnelMux) Demux() {

	// Demultiplexing action (route to proper channel)

	for {
		p, err := protocol.RecvPacket(h.ownerCon)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Extract id from the packet

		packetId := p.GetId()

		// Lookup & send to proper channel

		h.mu.Lock()
		c := h.subs[packetId]
		h.mu.Unlock()

		if c == nil {
			fmt.Printf("No channel for stream ID %d\n", packetId)
			continue
		}

		c <- p.GetPayload()
	}
}
