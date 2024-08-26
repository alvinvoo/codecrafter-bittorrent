package worker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/protocol"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/torrent"
)

type Downloader struct{}

func NewDownloader() *Downloader {
	return &Downloader{}
}

func (d *Downloader) Work(j Job) {
	dj, ok := j.(*DownloadPieceJob)
	if !ok {
		fmt.Println("Failed to type assert job to DownloadPieceJob")
		return
	}

	t := time.NewTimer(time.Duration(rand.Intn(5)) * time.Second)
	defer t.Stop()
	<-t.C
	fmt.Printf("Downloaded piece %d with length %d of %d \n", dj.PieceIndex, dj.Torrent.Info.Length, dj.Torrent.Info.PieceLength)
}

type DownloadPieceJob struct {
	Peers      []protocol.Peer // a peer list to rotate peers
	PieceIndex int
	Torrent    *torrent.TorrentMetadata
	Hash       string
}
