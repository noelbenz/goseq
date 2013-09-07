package goseq

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

const (
	tPlayersPacketReqID  byte = 0x55 // "U"
	tPlayersPacketRespID byte = 0x44 // "D"
	// Hint of the minimum bytes required per player.
	minPlayerSize int = 9
)

var (
	MissingPlayers error = errors.New("Packet contains less than reported amount of Players.")
)

// Player represents a Player currently on
// a given server.
type Player interface {
	// Index is the player's index in the server's
	// player list.
	Index() int
	// Name is the Steam Name of the player.
	Name() string
	// Score is the player's score as defined by the server.
	// Generally this is kills.
	Score() int
	// Duration is how long the player has been playing the server.
	Duration() time.Duration
}

// Wire format of a single player section.
type packetPtPlayer struct {
	_Index byte
	_Name  string
	Std2   struct {
		Score    int32
		Duration float32 // Number of seconds
	}
}

func (p packetPtPlayer) Index() int   { return int(p._Index) }
func (p packetPtPlayer) Name() string { return p._Name }
func (p packetPtPlayer) Score() int   { return int(p.Std2.Score) }
func (p packetPtPlayer) Duration() time.Duration {
	// We need to do floating point arithmetic so
	// we must convert both types to float64s
	// to lose the least precision.
	flrep := float64(p.Std2.Duration) * float64(time.Second)
	return time.Duration(flrep)
}

func (serv *iserver) Players(limit time.Duration) (players []Player, err error) {
	timer := time.NewTimer(limit)
	done := make(chan bool)

	go func() { players, err = serv.get_players(); done <- true }()

	select {
	case <-timer.C:
		return nil, Timeout
	case <-done:
		return
	}
}

func (serv *iserver) get_players() (players []Player, err error) {
	var challengeId int32
	if challengeId, err = serv.getChallenge(tPlayersPacketReqID); err != nil {
		return
	}

	conn, err := serv.getConnection()
	if err != nil {
		return
	}

	request := newWrappedChallengeBA(tPlayersPacketReqID, challengeId)
	if _, err = conn.Write(request); err != nil {
		return
	}

	stream := newPacketStream()
	if err = stream.Gobble(conn); err != nil {
		return
	}

	payload, err := stream.GetFullPayload()
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(payload)

	resp := wrPlayerResponse{}
	if err = binary.Read(buf, byteOrder, &resp); err != nil {
		return
	}

	if resp.Header != tPlayersPacketRespID {
		err = PacketMalformed
		return
	}

	if buf.Len() < minPlayerSize*int(resp.NumPlayers) {
		err = MissingPlayers
		return
	}

	for i := uint8(0); i < resp.NumPlayers; i++ {
		plr := packetPtPlayer{}
		// Index
		if err = binary.Read(buf, byteOrder, &plr._Index); err != nil {
			return
		}
		// Name
		if plr._Name, err = rcstr(buf); err != nil {
			return
		}
		// Frags & Kills
		if err = binary.Read(buf, byteOrder, &plr.Std2); err != nil {
			return
		}
		players = append(players, plr)
	}

	return
}

type wrPlayerResponse struct {
	Header     byte
	NumPlayers uint8
}
