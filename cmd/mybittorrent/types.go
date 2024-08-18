package main

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

type PeerResponse struct {
	Complete    int    `json:"complete"`
	Incomplete  int    `json:"incomplete"`
	Interval    int    `json:"interval"`
	MinInterval int    `json:"min interval"`
	Peers       []byte `json:"peers"`
}
