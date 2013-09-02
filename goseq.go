package goseq

import (
	"encoding/binary"
)

// Source byte order is little endian.
var byteOrder binary.ByteOrder = binary.LittleEndian

const packetHeaderSz int = 4

var (
	// packetHeader is the standard expected bytes at the head
	// of a packet
	packetHeader [packetHeaderSz]byte = [...]byte{0xFF, 0xFF, 0xFF, 0xFF}
)
