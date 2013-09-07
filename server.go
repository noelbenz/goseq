package goseq

import (
	"errors"
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
	Info(timeout time.Duration) (ServerInfo, error)
	// Players retrieves the players currently on a server.
	Players(timeout time.Duration) ([]Player, error)
	// Rules returns the server-defined rules of the server.
	// These are mostly Convar settings.
	Rules(timeout time.Duration) (RuleMap, error)
	SetAddress(string) error
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

// @TODO: Implement these
func (serv *iserver) Address() string                                 { return serv.addr }
func (serv *iserver) Players(timeout time.Duration) ([]Player, error) { return nil, nil }
func (s *iserver) SetAddress(a string) error                          { s.addr = a; s.remoteAddr = nil; return nil }

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

type wrChallengeResponse struct {
	Magic [4]byte
}
