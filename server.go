package goseq

import (
	"errors"
	"io"
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

// NewServer returns a Server that uses the network
// as its data source.
func NewServer() Server {
	return &iserver{
		src: &sourceRemote{},
	}
}

// implementation of Server
type iserver struct {
	src source
}

func (serv *iserver) Address() string        { return serv.src.address() }
func (s *iserver) SetAddress(a string) error { return s.src.setAddress(a) }

func (s *iserver) getConnection() (io.ReadWriteCloser, error) {
	return s.src.connection()
}

type wrChallengeResponse struct {
	Magic [4]byte
}
