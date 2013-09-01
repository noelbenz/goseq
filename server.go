package goseq

import (
	"time"
)

// Server represents a Source server.
type Server interface {
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
