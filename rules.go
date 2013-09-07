package goseq

import (
	"bytes"
	"encoding/binary"
	"time"
)

// Rules is a key-value pair of
// convar settings for the server.
type RuleMap map[string]string

func (serv *iserver) Rules(timeout time.Duration) (rmap RuleMap, err error) {
	rmap = newRuleMap()
	var challenge int32

	if challenge, err = serv.getChallenge(byte('V')); err != nil {
		return
	}

	conn, err := serv.getConnection()
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(packetHeader[0:])
	chal := wrRuleReq{Header: byte('V'), Challenge: challenge}

	if err = binary.Write(buf, byteOrder, chal); err != nil {
		return
	}

	chal_bytes := buf.Bytes()

	if _, err = conn.Write(chal_bytes); err != nil {
		return
	}

	st := newPacketStream()
	if err = st.Gobble(conn); err != nil {
		return
	}

	payload, err := st.GetFullPayload()
	if err != nil {
		return
	}

	buf = bytes.NewBuffer(payload)
	var rsp_wrapped struct {
		Header [4]byte
		wrRuleResponse
	}
	if err = binary.Read(buf, byteOrder, &rsp_wrapped); err != nil {
		return
	}

	rsp := rsp_wrapped.wrRuleResponse

	if rsp.Header != byte('E') {
		err = PacketHeaderErr
		return
	}

	// for each rule parse the string

	for n := int16(0); n < rsp.NumRules; n++ {
		var key string
		var val string

		key, err = rcstr(buf)
		if err != nil {
			return
		}
		val, err = rcstr(buf)
		if err != nil {
			return
		}

		rmap[key] = val
	}

	return
}

func newRuleMap() RuleMap {
	return make(RuleMap)
}

type wrRuleReq struct {
	Header    byte
	Challenge int32
}

type wrRuleResponse struct {
	Header   byte
	NumRules int16
}
