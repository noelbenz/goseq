package main

import (
    "io"
    "bufio"
)

type PingResponse struct {
    header  byte
}

func (r *PingResponse) Read(stream io.Reader) error {
    header, err := NewReader(PingResponse).ReadByte()
    return err
}
