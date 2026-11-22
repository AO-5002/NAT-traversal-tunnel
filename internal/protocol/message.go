package protocol

import (
	"encoding/binary"
	"io"
)

type Packet struct {
	id            uint32
	payloadLength uint32
	payload       []byte
}

func (p *Packet) GetPayload() []byte {
	return p.payload
}

func (p *Packet) GetId() uint32 {
	return p.id
}

func NewPacket(streamID uint32, payload []byte) *Packet {
	return &Packet{
		id:            streamID,
		payloadLength: uint32(len(payload)),
		payload:       payload,
	}
}

// General Protocol Recv/Send

func RecvPacket(r io.Reader) (*Packet, error) {

	// Create the header & read into it

	hdr := make([]byte, 8)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return nil, err
	}

	// Create the complete payload

	length := binary.BigEndian.Uint32(hdr[4:8])
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	// Return
	return &Packet{
		id:      binary.BigEndian.Uint32(hdr[:4]),
		payload: payload,
	}, nil
}

func SendPacket(p *Packet, w io.Writer) error {

	// Encode the packet

	buf := make([]byte, 0, 8+len(p.payload))
	buf = binary.BigEndian.AppendUint32(buf, p.id)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(p.payload)))
	buf = append(buf, p.payload...)

	// Send the packet to the proper connection

	_, err := w.Write(buf)
	return err
}
