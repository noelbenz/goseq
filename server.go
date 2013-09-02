package goseq

import (
	"errors"
	"io"
	"net"
	"time"
)

// ServerType is the type of server being queried.
type ServerType byte

const (
	Dedicated ServerType = ServerType(byte('D'))
	Listen    ServerType = ServerType(byte('L')) // "Non dedicated"
	SourceTV  ServerType = ServerType(byte('P'))
)

// ServerEnvironment is the OS the server is running on.
type ServerEnvironment byte

const (
	Linux   ServerEnvironment = ServerEnvironment(byte('L'))
	Windows ServerEnvironment = ServerEnvironment(byte('W'))
)

const (
	// PayloadSize is the official packet-payload size
	// of UDP headers.
	PayloadSize int = 1400
)

const (
	// NoAddress represents a address that has yet to be
	// set by the server implementation.
	NoAddress string = "address-not-set.example.org:0"
)

var (
	NoAddressSet error = errors.New("The server does not have a remote address set.")
)

// Server represents a Source server.
type Server interface {
	Address() string
	// Ping returns the connection latency of a server.
	Ping(timeout time.Duration) (time.Duration, error)
	// Info returns the Source Info query result.
	Info() (map[string]interface{}, error)
	// Players retrieves the players currently on a server.
	Players() ([]Player, error)
	// Rules returns the server-defined rules of the server.
	// These are mostly Convar settings.
	Rules() ([]Rule, error)
}

func NewServer() Server {
	return &iserver{
		addr: NoAddress,
	}
}

// implementation of Server
type iserver struct {
	addr       string
	remoteAddr *net.UDPAddr
}

func (s *iserver) setAddress(a string) { s.addr = a; s.remoteAddr = nil }

// @TODO: Implement these
func (serv *iserver) Address() string                       { return serv.addr }
func (serv *iserver) Info() (map[string]interface{}, error) { return make(map[string]interface{}), nil }
func (serv *iserver) Players() ([]Player, error)            { return nil, nil }
func (serv *iserver) Rules() ([]Rule, error)                { return nil, nil }

func (s *iserver) getConnection() (*net.UDPConn, error) {
	if s.Address() == NoAddress {
		return nil, NoAddressSet
	}
	if s.remoteAddr == nil {
		var err error
		s.remoteAddr, err = net.ResolveUDPAddr("udp", s.addr)
		if err != nil {
			s.remoteAddr = nil
			return nil, err
		}
	}
	conn, err := net.DialUDP("udp", nil, s.remoteAddr)
	if err != nil {
		s.remoteAddr = nil
		return nil, err
	}
	return conn, nil
}

// binary form of the server packet
type packetServer struct {
	Header struct {
		Magic       [packetHeaderSz]byte
		Designation byte // 0x49 "I"
	}
	// always present
	Std1 struct {
		Header      byte
		Protocol    byte
		Name        string
		Map         string
		Folder      string
		Game        string
		ID          int16 // App ID of game
		Players     uint8
		MaxPlayers  uint8
		Bots        uint8
		Servertype  ServerType
		Environment byte
		Visibility  byte
		VAC         bool
	}
	// only if the game is The Ship (ID == 2400)
	Ship struct {
		Mode      byte
		Witnesses uint8
		Duration  uint8
	}
	// always present after the ship
	Std2 struct {
		Version string
		EDF     byte // Extra Data Flag
	}
	// fields that depend on the flags in EDF
	Extra struct {
		Port          uint16 // if EDF & 0x80
		SteamID       uint64 // if EDF & 0x10
		SpectatorPort uint16 // if EDF & 0x40
		SpectatorName string // if EDF & 0x40
		Keywords      string // if EDF & 0x20
		GameID        uint64 // if  EDF & 0x01
	}
}

func (_ *packetServer) Identifier() byte {
	return 0x49 // "I"
}

func (p *packetServer) Decode(stream io.Reader) error {
	//@TODO(hunter): implement :)
	return nil
}
