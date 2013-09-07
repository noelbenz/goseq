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

type wrappedChallenge struct {
	Magic [4]byte
	challenge
}

type challenge struct {
	Type      byte
	Challenge int32
}

// Utility function to wrap challenges and turn it into
// a byte array for sending.
func newWrappedChallengeBA(chalType byte, chalValue int32) []byte {
	chal := wrappedChallenge{
		Magic: [4]byte{0xFF, 0xFF, 0xFF, 0xFF},
		challenge: challenge{
			Type:      chalType,
			Challenge: chalValue,
		},
	}

	buf := bytes.NewBuffer(make([]byte, 0, 9))
	binary.Write(buf, byteOrder, chal)
	return buf.Bytes()
}
