package goseq

import (
	"io"
)

// packet represents a packet
// from the wire
type packet interface {
	// The constant used to
	Identifier() byte
	// decode to a packet from a stream
	Decode(io.Reader) error
}
