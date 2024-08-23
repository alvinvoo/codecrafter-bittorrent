package main

import (
	"crypto/sha1"
	"encoding/binary"
)

type TorrentMetadata struct {
	Announce string   `json:"announce"`
	Info     InfoDict `json:"info"`
}

type InfoDict struct {
	Length      int    `json:"length"`
	Name        string `json:"name"`
	PieceLength int    `json:"piece length"`
	Pieces      []byte `json:"pieces"`
}

func (info InfoDict) hash() []byte {
	encodeInfoDict := encodeInfoDict(info)
	h := sha1.New()
	h.Write([]byte(encodeInfoDict))
	return h.Sum(nil)
}

type Handshake struct {
	length   byte
	protocol string
	resv     [8]byte
	info     []byte
	peerId   []byte
}

func (handshake Handshake) encode() []byte {
	var msg []byte
	msg = append(msg, handshake.length)
	msg = append(msg, handshake.protocol...)
	msg = append(msg, handshake.resv[:]...)
	msg = append(msg, handshake.info...)
	msg = append(msg, handshake.peerId...)
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
