package multiplexing

import (
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

func (h *TunnelMux) Mux(id uint32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Multiplexing action

}

func (h *TunnelMux) Demux(con net.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Demultiplexing action
}
