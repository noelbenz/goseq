package goseq

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var ChallengeFailed error = errors.New("Could not retrieve challenge.")
var ChallengeSizeMismatch error = errors.New("Challenge was the wrong size.")

func (s *iserver) getChallenge(rqtype byte) (ch int32, err error) {

	buf := bytes.NewBuffer([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	binary.Write(buf, byteOrder, challenge{
		Type:      rqtype,
		Challenge: -1,
	})

	conn, err := s.getConnection()
	if err != nil {
		return
	}

	chRequest := buf.Bytes()

	if _, err = conn.Write(chRequest); err != nil {
		return
	}

	st := newPacketStream()
	if err = st.Gobble(conn); err != nil {
		return
	}

	payload, err := st.GetFullPayload()
	buf = bytes.NewBuffer(payload)

	chal := challenge{}
	if err = binary.Read(buf, byteOrder, &chal); err != nil {
		return
	}

	if chal.Type != 0x41 {
		err = PacketMalformed
		return
	}

	ch = chal.Challenge
	return
}

type challenge struct {
	Type      byte
	Challenge int32
}
