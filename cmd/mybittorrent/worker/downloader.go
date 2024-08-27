package worker

import (
	"fmt"
	"sync"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/protocol"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/torrent"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/util"
)

type Downloader struct {
	Peers    []protocol.Peer
	FullData []byte
	mu       sync.Mutex
}

func NewDownloader(peers []protocol.Peer, length int) *Downloader {
	return &Downloader{
		Peers:    peers,
		FullData: make([]byte, length),
	}
}

func (d *Downloader) CloseConnections() {
	for _, p := range d.Peers {
		p.Conn.Close()
	}
}

func (d *Downloader) Work(j Job) {
	dj, ok := j.(*DownloadPieceJob)
	if !ok {
		fmt.Println("Failed to type assert job to DownloadPieceJob")
		return
	}

	// first assign a peer to download the piece
	// TODO: implement a better way to select a peer
	p := d.Peers[dj.PieceIndex%len(d.Peers)]

	// initialize peer if not already
	if !p.Init {
		err := protocol.DownloadInit(p.Conn)
		if err != nil {
			fmt.Println("Failed to initialize peer:", err)
			return
		}
	}

	// request the piece
	piece := protocol.RequestPiece(p.Conn, dj.Torrent, dj.PieceIndex)

	// seems like redundant
	d.mu.Lock()
	defer d.mu.Unlock()

	copiedLen := copy(d.FullData[dj.PieceIndex*dj.Torrent.Info.PieceLength:], piece)
	util.DebugLog(fmt.Sprintf("Copied %d bytes of piece %d", copiedLen, dj.PieceIndex))
	if copiedLen != len(piece) {
		fmt.Println("Failed to copy the piece")
		// TODO: handle failure gracefully
		dj.Failed = true
		return
	}

	dj.completed = true
}

type DownloadPieceJob struct {
	PieceIndex int
	Torrent    *torrent.TorrentMetadata
	Hash       string
	completed  bool
	Failed     bool
}
