package goseq

import (
	"time"
)

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

// Wire format of the player.
type packedPlayer struct {
	index    byte
	name     string
	score    int32
	duration float32 // Number of seconds
}

func (p *packedPlayer) Index() int { return int(p.index) }
func (p *packedPlayer) Name() int  { return p.name }
func (p *packedPlayer) Score() int { return int(p.score) }
func (p *packedPlayer) Duration() int {
	return time.Duration(p.duration * time.Second)
}
