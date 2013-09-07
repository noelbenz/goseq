package goseq

import (
	"encoding/binary"
	"errors"
	"time"
)

var (
	Timeout error = errors.New("The server did not respond in time (timeout).")
)

func (s *iserver) Ping(timeout time.Duration) (time.Duration, error) {
	conn, err := s.getConnection()
	if err != nil {
		return 0, err
	}

	timer := make(chan bool)
	done := make(chan time.Duration)

	request := append(packetHeader[0:], 0x69)
	response := make([]byte, binary.Size(pingPacket{}))
	n := 0

	go func() {
		time.Sleep(timeout)
		timer <- true
	}()
	go func() {

		conn.Write(request[0:])
		start := time.Now()
		n, err = conn.Read(response[0:])
		done <- time.Now().Sub(start)
	}()

	select {
	case <-timer:
		return 0, Timeout
	case took := <-done:
		return took, err
	}
}

type pingPacket struct {
	Header struct {
		Magic       [packetHeaderSz]byte
		Designation byte // 0x6A "j"
	}
	Payload [14]byte // just filled with 0's (0x30)
	Footer  byte     // 0x0
}
