package goseq

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"
)

// Region is a part of the world in which
// Source servers are segregated. They should
// have made these flags so you query multiple
// regions at once. But alas, they did not.
// Let's hope it ends up in Source II.
type Region byte

const (
	USEast       Region = 0x00
	USWest              = 0x01
	SouthAmerica        = 0x02
	Europe              = 0x03
	Asia                = 0x04
	Australia           = 0x05
	MiddleEast          = 0x06
	Africa              = 0x07
	RestOfWorld         = 0xFF
)

const (
	// Beggining the the special ip address that Master
	// interprets as the beggining and end of an IP listing.
	Beggining string = "0.0.0.0:0"
)

const (
	masterRespHeaderLength int = 6
)

var (
	// MasterSourceServers are the ip addresses of the Master servers for all Source engine servers.
	MasterSourceServers []string = []string{
		"68.177.101.62:27011",
		"69.28.158.131:27011",
		"208.64.200.117:27011",
		"208.64.200.118:27011",
	}
	// which server we're going to use by default
	favored_server int = 2

	MasterServerTimeout time.Duration = 5 * time.Second
)

var (
	masterResponseHeader [masterRespHeaderLength]byte = [...]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x66, 0x0A}
)

type MasterServer interface {
	SetFilter(Filter) error
	GetFilter() Filter
	SetAddr(string) error
	GetAddr() string
	SetRegion(Region)
	GetRegion() Region
	// Query returns results starting at the server
	// with startIP.
	// Returned servers are NOT guaranteed to work.
	Query(startIP string) ([]Server, error)
}

type master struct {
	filter       Filter
	addr         string
	master_index int
	region       Region
	remoteAddr   *net.UDPAddr
	remoteConn   *net.UDPConn
}

func NewMasterServer() MasterServer {
	return &master{
		filter:       NewFilter(),
		addr:         MasterSourceServers[favored_server],
		master_index: favored_server,
		region:       USWest,
		remoteAddr:   nil,
		remoteConn:   nil,
	}
}

func (m *master) SetFilter(f Filter) error { m.filter = f; return nil }
func (m *master) GetFilter() Filter        { return m.filter }
func (m *master) SetAddr(i string) error   { m.addr = i; m.remoteAddr = nil; return nil }
func (m *master) GetAddr() string          { return m.addr }
func (m *master) SetRegion(i Region)       { m.region = i }
func (m *master) GetRegion() Region        { return m.region }

func (m *master) refreshConnection() (err error) {
	if m.remoteAddr == nil || m.remoteConn == nil {
		m.remoteAddr, err = net.ResolveUDPAddr("udp", m.addr)
		if err != nil {
			m.remoteAddr = nil
			return err
		}
		m.remoteConn, err = net.DialUDP("udp", nil, m.remoteAddr)
		if err != nil {
			m.remoteAddr = nil
			return err
		}
	}
	return nil
}

func (m *master) makerequest(ip string) []byte {
	packet := bytes.NewBuffer([]byte{})
	packet.WriteByte(0x31)
	packet.WriteByte(byte(m.region))
	packet.WriteString(ip)
	packet.WriteByte(0x0)
	packet.WriteString(string(m.filter.GetFilterFormat()))

	req := packet.Bytes()
	return req
}

// performs no allocations to keep it fast
// iterating over hundreds of servers.
func (_ *master) ip2server(ip wireIP, serv *iserver) {
	serv.src.setAddress(ip.String())
}

// try to write and read from the socket.
func (m *master) try(request, buffer []byte) (error, int) {
	timeout := make(chan bool, 1)
	done := make(chan error, 1)
	n := 0
	var e error

	go func() {
		if e := m.refreshConnection(); e != nil {
			done <- e
			return
		}

		if _, e := m.remoteConn.Write(request); e != nil {
			done <- e
			return
		}

		n, e = m.remoteConn.Read(buffer)
		if e != nil {
			done <- e
			return
		}
		done <- nil
	}()

	go func() {
		time.Sleep(MasterServerTimeout)
		timeout <- true
	}()

	select {
	case e := <-done:
		return e, n
	case <-timeout:
		return Timeout, 0
	}
}

func (m *master) Query(at string) ([]Server, error) {
	reqpacket := m.makerequest(at)
	respbuffer := [1024 * 1024 * 2]byte{} // 2MB, should be 400 bytes above max

	var e error
	var n int

	start_indice := m.master_index
	for {
		e, n = m.try(reqpacket, respbuffer[0:])
		if e == Timeout {
			m.master_index = (m.master_index + 1) % len(MasterSourceServers)
			m.addr = MasterSourceServers[m.master_index]

			if m.master_index == start_indice {
				// we've come full circle, time to quit.
				return nil, Timeout
			}

			favored_server = m.master_index
			m.remoteAddr = nil
		} else if e != nil {
			return nil, e
		} else {
			break
		}
	}

	resp := wireMasterResponse{}
	err := resp.Decode(bytes.NewBuffer(respbuffer[0:n]), n)
	if err != nil {
		return nil, err
	}

	servers := make([]Server, len(resp.Ips))

	var iterated_server Server
	for i, ip := range resp.Ips {
		iterated_server = NewServer()
		iterated_server.SetAddress(ip.String())
		servers[i] = iterated_server
	}

	return servers, nil
}

// Incoming IPs as represented on the wire.
type wireIP struct {
	Oct struct {
		O1,
		O2,
		O3,
		O4 byte
	}
	// ATCHTUNG!!! This is NETWORK BYTE ORDERED
	// as defined by the spec.
	Port uint16
}

func (p wireIP) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d",
		p.Oct.O1, p.Oct.O2, p.Oct.O3, p.Oct.O4, p.Port)
}

type wireMasterResponse struct {
	Head struct {
		Magic [masterRespHeaderLength]byte
	}
	Ips []wireIP
}

// Decode fills the packet withinformation of the packet.
func (r *wireMasterResponse) Decode(packet io.Reader, n int) error {
	if r.Ips == nil {
		r.Ips = make([]wireIP, 0, 50)
	}

	err := binary.Read(packet, byteOrder, &r.Head)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(r.Head.Magic, masterResponseHeader) {
		return errors.New("Header does not match.")
	}

	remaining := n - binary.Size(r.Head.Magic)
	ipsize := binary.Size(wireIP{})

	for ; remaining >= ipsize; remaining -= ipsize {
		ip := wireIP{}
		// Normal little endian read.
		if err := binary.Read(packet, byteOrder, &ip.Oct); err != nil {
			return err
		}
		// Seperate read because of big endian requirement
		if err := binary.Read(packet, binary.BigEndian, &ip.Port); err != nil {
			return err
		}
		r.Ips = append(r.Ips, ip)
	}

	return nil
}
