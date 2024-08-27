package protocol

import (
	"encoding/binary"
	"net"
)

type Handshake struct {
	length   byte
	protocol string
	resv     [8]byte
	info     []byte
	PeerId   []byte
}

func (handshake Handshake) encode() []byte {
	var msg []byte
	msg = append(msg, handshake.length)
	msg = append(msg, handshake.protocol...)
	msg = append(msg, handshake.resv[:]...)
	msg = append(msg, handshake.info...)
	msg = append(msg, handshake.PeerId...)
	return msg
}

type PeerResponse struct {
	Complete    int    `json:"complete"`
	Incomplete  int    `json:"incomplete"`
	Interval    int    `json:"interval"`
	MinInterval int    `json:"min interval"`
	Peers       []byte `json:"peers"`
}

type PeerMessage struct {
	Length uint32
	Id     uint8
	index  uint32
	begin  uint32
	block  uint32
}

func (peerMessage PeerMessage) encode() []byte {
	var msg []byte
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, peerMessage.Length)
	msg = append(msg, lengthBytes...)
	msg = append(msg, peerMessage.Id)
	return msg
}

type Peer struct {
	Conn  *net.TCPConn // need to close at the very end
	Id    string
	Retry int
	Init  bool
}
