package goseq

import (
	"bufio"
	"io"
)

type PingResponse struct {
	header byte
}

func (r *PingResponse) Read(stream io.Reader) error {
	header, err := NewReader(PingResponse).ReadByte()
	return err
}
