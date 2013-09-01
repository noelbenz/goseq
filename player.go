package goseq

import (
	"time"
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
	index    byte
	name     string
	score    int32
	duration float32 // Number of seconds
}

func (p *packetPtPlayer) Index() int   { return int(p.index) }
func (p *packetPtPlayer) Name() string { return p.name }
func (p *packetPtPlayer) Score() int   { return int(p.score) }
func (p *packetPtPlayer) Duration() time.Duration {
	// We need to do floating point arithmetic so
	// we must convert both types to float64s
	// to lose the least precision.
	flrep := float64(p.duration) * float64(time.Second)
	return time.Duration(flrep)
}
