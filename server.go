package goseq

import (
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

// implementation of Server
type iserver struct {
	addr string
}

// @TODO: Implement these
func (serv iserver) Address() string                                   { return serv.addr }
func (serv iserver) Ping(timeout time.Duration) (time.Duration, error) { return time.Duration(0), nil }
func (serv iserver) Info() (map[string]interface{}, error)             { return make(map[string]interface{}), nil }
func (serv iserver) Players() ([]Player, error)                        { return nil, nil }
func (serv iserver) Rules() ([]Rule, error)                            { return nil, nil }

// binary form of the server packet
type packetServer struct {
	// always present
	std1 struct {
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
	ship struct {
		Mode      byte
		Witnesses uint8
		Duration  uint8
	}
	// always present after the ship
	std2 struct {
		Version string
		EDF     byte // Extra Data Flag
	}
	// fields that depend on the flags in EDF
	extra struct {
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
