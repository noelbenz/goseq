package goseq

import (
	"io"
	"net"
)

// source allows us to mock a
// connection to a server.
type source interface {
	address() string
	setAddress(string) error
	connection() (io.ReadWriteCloser, error)
}

// sourceRemote is a connection
// to a foreign server.
type sourceRemote struct {
	ip     string
	remote *net.UDPAddr
}

func (src *sourceRemote) address() string { return src.ip }

func (s *sourceRemote) setAddress(a string) error {
	s.ip = a
	s.remote = nil
	return nil
}

func (s *sourceRemote) connection() (io.ReadWriteCloser, error) {
	if s.ip == NoAddress || s.ip == "" {
		return nil, NoAddressSet
	}
	if s.remote == nil {
		var err error
		s.remote, err = net.ResolveUDPAddr("udp", s.ip)
		if err != nil {
			s.remote = nil
			return nil, err
		}
	}
	conn, err := net.DialUDP("udp", nil, s.remote)
	if err != nil {
		s.remote = nil
		return nil, err
	}
	return conn, nil
}
