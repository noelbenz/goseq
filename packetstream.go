package goseq

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const (
	pkt_NOT_SPLIT uint32 = 0xFFFFFFFF
	pkt_SPLIT     uint32 = 0xFFFFFFFE
)

const (
	// PayloadSize is the official packet-payload size
	// of UDP headers.
	PayloadSize int = 1400
)

var (
	PacketMalformed     error = errors.New("Packet appears malformed.")
	PayloadSizeMismatch error = errors.New("Decompressed payload is not expected size.")
	PayloadCRC32Fail    error = errors.New("CRC validation failed on decompressed payload. Possible corruption?")
)

type pkt_header struct {
	Std struct {
		HeaderCode uint32
	}
	Extended struct { // if HeaderCode == pkt_SPLIT
		Std struct {
			ID     uint32
			Total  byte
			Number byte
			Size   int16
		}
		ComprInf struct { // if ID & (1 << 31)
			Size  uint32
			CRC32 uint32
		}
	}
}

type packet struct {
	Header  pkt_header
	Payload []byte
}

func newPacket() packet {
	header := pkt_header{}
	header.Extended.Std.Number = 0
	header.Extended.Std.Total = 1
	return packet{
		Payload: make([]byte, 0, PayloadSize),
		Header:  header,
	}
}

// Reads the packets coming from the wire
// and turns them into a single payload.
// Strips all Source Packet specific headers.
type packetStream struct {
	expected int
	packets  []packet
}

func newPacketStream() packetStream {
	return packetStream{
		expected: 1,
		packets:  make([]packet, 1),
	}
}

func pk_signals_compression(header *pkt_header) bool {
	return (header.Extended.Std.ID & (1 << 31)) > 0
}

// Take a byte buffer and assume it's a fuill UDP packet,
// construct Source packet from it. AKA interpret source
// headers and find where the payload starts.
func contructPacket(full []byte) (packet, error) {
	pk := newPacket()
	buf := bytes.NewBuffer(full)
	var err error

	if err = binary.Read(buf, byteOrder, &pk.Header.Std.HeaderCode); err != nil {
		return pk, err
	}

	if pk.Header.Std.HeaderCode == pkt_NOT_SPLIT {
		pk.Payload = buf.Bytes()
		return pk, nil
	}

	// weed out any funny packets that
	// aren't correct
	if pk.Header.Std.HeaderCode != pkt_SPLIT {
		return pk, PacketMalformed
	}

	// The packet is not assumed to be split.
	// Read split headers.
	if err = binary.Read(buf, byteOrder, &pk.Header.Extended.Std); err != nil {
		return pk, err
	}

	// test for compression
	if pk_signals_compression(&pk.Header) {
		// decode compression information
		if err = binary.Read(buf, byteOrder, &pk.Header.Extended.ComprInf); err != nil {
			return pk, err
		}

		// note that the payload is actually compressed and should
		// be decompressed during read.
	}

	// Payload is the rest!
	pk.Payload = buf.Bytes()

	return pk, nil
}

// Gobble packets from the connection.
// Number of packets are determined by the format of
// the packet.
func (st *packetStream) Gobble(reader io.Reader) error {
	for got := 0; got < st.expected; got++ {
		var buffer [PayloadSize]byte
		var n int
		var err error
		var pk packet

		if n, err = reader.Read(buffer[0:]); err != nil {
			return err
		}

		if pk, err = contructPacket(buffer[0:n]); err != nil {
			return err
		}

		// resize packet buffer to fit new expected total
		if st.expected != int(pk.Header.Extended.Std.Total) {
			st.expected = int(pk.Header.Extended.Std.Total)
			st.packets = make([]packet, st.expected)
		}

		// noone like buffer overflows.
		if int(pk.Header.Extended.Std.Number) >= len(st.packets) {
			// a naughty server is trying to crash us.
			return PacketMalformed
		}

		st.packets[pk.Header.Extended.Std.Number] = pk
	}
	return nil
}

// Returns the conitguous payload, decompressing if necessary.
func (st *packetStream) GetFullPayload() ([]byte, error) {
	payload := st.contiguous_payload()

	// token packet header to decide if to decompress
	pkh := &st.packets[0].Header

	if !pk_signals_compression(pkh) {
		// no need to decompress, just return the payload
		return payload, nil
	}

	// Decompressed buffer
	decompressed := bytes.NewBuffer(make([]byte, 0, pkh.Extended.ComprInf.Size))
	// Decompress data
	copied, err := io.Copy(decompressed, bzip2.NewReader(bytes.NewBuffer(payload)))

	if err != nil {
		return nil, err
	}

	if copied != int64(pkh.Extended.ComprInf.Size) {
		return nil, PayloadSizeMismatch
	}

	decompressed_buf := decompressed.Bytes()

	// time to checksum dis place up
	checksum := crc32.ChecksumIEEE(decompressed_buf)

	if checksum != pkh.Extended.ComprInf.CRC32 {
		// all that work for nothing :(
		return nil, PayloadCRC32Fail
	}

	return decompressed_buf, nil
}

func (st *packetStream) contiguous_payload() []byte {
	sum := 0
	for _, pk := range st.packets {
		sum += len(pk.Payload)
	}

	contiguous := make([]byte, sum)

	sum = 0
	for _, pk := range st.packets {
		copy(contiguous[sum:], pk.Payload)
		sum += len(pk.Payload)
	}

	return contiguous
}
