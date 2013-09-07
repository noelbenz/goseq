package goseq

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	HAS_PORT     byte = 0x80
	HAS_STEAMID       = 0x10
	HAS_SOURCETV      = 0x40
	HAS_KEYWORDS      = 0x20
	HAS_GAMEID        = 0x01
)

func (serv *iserver) Info(timeout time.Duration) (info ServerInfo, err error) {
	info = NewServerInfo()

	conn, err := serv.getConnection()
	if err != nil {
		return info, err
	}
	defer conn.Close()

	outOfTime := make(chan bool)
	done := make(chan bool)

	request := []byte("\xFF\xFF\xFF\xFF\x54Source Engine Query\x00")
	// a buffer of buffers for multi packet responses
	var payload []byte

	go func() {
		time.Sleep(timeout)
		outOfTime <- true
	}()

	go func() {
		conn.Write(request)
		pks := newPacketStream()
		err = pks.Gobble(conn) // yum yum
		if err != nil {
			done <- true
		}
		payload, err = pks.GetFullPayload()
		done <- true
	}()

	select {
	case <-outOfTime:
		return info, Timeout
	case <-done:
		if err != nil {
			return info, err
		}
		break
	}

	err = info.decode(bytes.NewBuffer(payload))
	return
}

type wrInfHead struct {
	Header byte // 0x49 "I"
}

type wrInfStd1 struct {
	Protocol byte
}

type wrInfStd2 struct {
	Name   string
	Map    string
	Folder string
	Game   string
}

type wrInfStd3 struct {
	ID          int16 // App ID of game
	Players     uint8
	MaxPlayers  uint8
	Bots        uint8
	Servertype  ServerType
	Environment byte
	Visibility  byte
	VAC         byte // is bool
}

type wrInfShip struct {
	Mode      byte
	Witnesses uint8
	Duration  uint8
}
type wrInfStd4 struct {
	Version string
	EDF     byte // Extra Data Flag
}
type wrInfExtra struct {
	Port          uint16 // if EDF & 0x80
	SteamID       uint64 // if EDF & 0x10
	SpectatorPort uint16 // if EDF & 0x40
	SpectatorName string // if EDF & 0x40
	Keywords      string // if EDF & 0x20
	GameID        uint64 // if EDF & 0x01
}

// ServerInfo holds all the info values for a server.
// Why is it composed so weird? Composition allows incrementally
// decoding it from the payload while still allowing external
// packages to access the data. AKA it hides the dirty-nasty.
//
// Eg ServerInfo{}.Port
type ServerInfo struct {
	wrInfHead
	// always present
	wrInfStd1
	wrInfStd2
	wrInfStd3
	// only if the game is The Ship (ID == 2400)
	wrInfShip
	wrInfStd4
	// fields that depend on the flags in EDF
	wrInfExtra
}

func NewServerInfo() ServerInfo {
	return ServerInfo{}
}

func (p *ServerInfo) decode(stream *bytes.Buffer) (err error) {
	if err = binary.Read(stream, byteOrder, &p.wrInfHead); err != nil {
		return
	}

	if p.wrInfHead.Header != byte('I') {
		return PacketMalformed
	}

	if err = binary.Read(stream, byteOrder, &p.wrInfStd1); err != nil {
		return
	}

	str2decode := []*string{&p.Name, &p.Map, &p.Folder, &p.Game}

	for _, loc := range str2decode {
		if *loc, err = rcstr(stream); err != nil {
			return
		}
	}

	if err = binary.Read(stream, byteOrder, &p.wrInfStd3); err != nil {
		return
	}

	// The Ship-only section
	if p.ID == 2400 {
		if err = binary.Read(stream, byteOrder, &p.wrInfShip); err != nil {
			return
		}
	}

	// Back to standard sections
	if p.Version, err = rcstr(stream); err != nil {
		return
	}

	if err = binary.Read(stream, byteOrder, &p.EDF); err != nil {
		return
	}

	// No to check all the extra flags! :D
	if p.EDF&HAS_PORT > 0 {
		if err = binary.Read(stream, byteOrder, &p.Port); err != nil {
			return
		}
	}
	if p.EDF&HAS_STEAMID > 0 {
		if err = binary.Read(stream, byteOrder, &p.SteamID); err != nil {
			return
		}
	}
	if p.EDF&HAS_SOURCETV > 0 {
		if err = binary.Read(stream, byteOrder, &p.SpectatorPort); err != nil {
			return
		}
		if p.Version, err = rcstr(stream); err != nil {
			return
		}
	}
	if p.EDF&HAS_KEYWORDS > 0 {
		if p.Keywords, err = rcstr(stream); err != nil {
			return
		}
	}
	if p.EDF&HAS_GAMEID > 0 {
		if err = binary.Read(stream, byteOrder, &p.GameID); err != nil {
			return
		}
	}
	return nil
}
