package torrent

import (
	"crypto/sha1"
	"fmt"
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

func (info InfoDict) Hash() []byte {
	encodeInfoDict := EncodeInfoDict(info)
	h := sha1.New()
	h.Write([]byte(encodeInfoDict))
	return h.Sum(nil)
}

func EncodeInfoDict(info InfoDict) string {
	return fmt.Sprintf("d6:lengthi%de4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
		info.Length, len(info.Name), info.Name, info.PieceLength, len(info.Pieces), info.Pieces)
}
