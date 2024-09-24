package main

import (
	// Uncomment this line to pass the first stage
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/protocol"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/torrent"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/util"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/worker"
)

func decodeFile(fileName string) (torrent.TorrentMetadata, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return torrent.TorrentMetadata{}, tracerr.Wrap(err)
	}

	metadataMap, rest, err := bencode.DecodeBencode(content)
	if err != nil {
		fmt.Println("bencode.DecodeBencode error:", err)
		return torrent.TorrentMetadata{}, tracerr.Wrap(err)
	}

	if len(rest) != 0 {
		fmt.Println("Rest is not empty. Invalid syntax")
		return torrent.TorrentMetadata{}, tracerr.Wrap(err)
	}

	// Type assertion to convert interface{} to map[string]interface{}
	decodedMap, ok := metadataMap.(map[string]interface{})
	if !ok {
		fmt.Println("Failed to type assert metadataMap to map[string]interface{}")
		return torrent.TorrentMetadata{}, tracerr.Wrap(err)
	}

	// Convert the decoded map to the torrentMetadata.TorrentMetadata struct
	var torrentMetadata torrent.TorrentMetadata
	if announce, ok := decodedMap["announce"].(string); ok {
		torrentMetadata.Announce = string(announce)
	}

	if infoMap, ok := decodedMap["info"].(map[string]interface{}); ok {
		var infoDict torrent.InfoDict
		if length, ok := infoMap["length"].(int); ok {
			infoDict.Length = length
		}
		if name, ok := infoMap["name"].(string); ok {
			infoDict.Name = string(name)
		}
		if pieceLength, ok := infoMap["piece length"].(int); ok {
			infoDict.PieceLength = pieceLength
		}
		if pieces, ok := infoMap["pieces"].([]byte); ok {
			infoDict.Pieces = pieces // pieces are non-UTF-8 bytes
		}
		torrentMetadata.Info = infoDict
	}

	return torrentMetadata, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, rest, err := bencode.DecodeBencode([]byte(bencodedValue))
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		if len(rest) != 0 {
			fmt.Println("Rest is not empty. Invalid syntax")
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		fileName := os.Args[2]

		torrent, err := decodeFile(fileName)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		// Now you can use the struct
		fmt.Printf("Tracker URL: %s\n", torrent.Announce)
		fmt.Printf("Length: %d\n", torrent.Info.Length)
		fmt.Printf("Info Hash: %s\n", bencode.CalculateInfoHash(torrent))
		fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
		fmt.Printf("Piece Hashes:\n")
		pieces := bencode.SplitPiecesIntoHashes(torrent.Info.Pieces)
		for _, p := range pieces {
			fmt.Printf("%s\n", p)
		}
	} else if command == "peers" {
		fileName := os.Args[2]

		torrent, err := decodeFile(fileName)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		peersList, err := protocol.GetPeers(torrent)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		for _, p := range peersList {
			fmt.Println(p)
		}
	} else if command == "handshake" {
		fileName := os.Args[2]

		torrent, err := decodeFile(fileName)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		peerIpPort := os.Args[3]

		conn := protocol.EstablishTCPConnection(peerIpPort)
		defer conn.Close()

		response := protocol.SendTCPHandshake(conn, torrent)
		handshake := protocol.DestructureHandshakeResponse(response)

		fmt.Printf("Peer ID: %x\n", string(handshake.PeerId))
	} else if command == "download_piece" {
		option := os.Args[2]
		if option == "-o" && len(os.Args) == 6 {
			filePath := os.Args[3]
			fileName := os.Args[4]
			pieceIndexToDownload, err := strconv.Atoi(os.Args[5])
			if err != nil {
				fmt.Println("Invalid piece index")
				return
			}

			torrent, err := decodeFile(fileName)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			pieces := bencode.SplitPiecesIntoHashes(torrent.Info.Pieces)
			if (pieceIndexToDownload >= len(pieces)) || (pieceIndexToDownload < 0) {
				fmt.Println("Invalid piece index")
				return
			}

			peersList, err := protocol.GetPeers(torrent)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			// just use the first peer, since there's no specification
			conn := protocol.EstablishTCPConnection(peersList[0])
			defer conn.Close()

			response := protocol.SendTCPHandshake(conn, torrent)
			if (response == nil) || len(response) == 0 || (conn == nil) {
				fmt.Println("Error sending handshake")
				return
			}
			data := protocol.DownloadPiece(conn, torrent, pieceIndexToDownload)
			if util.GenerateSHA1Checksum(data) != pieces[pieceIndexToDownload] {
				fmt.Printf("Sha1 Checksum for Piece %d does not match\n", pieceIndexToDownload)
				return
			}

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Printf("Piece downloaded to %s.\n", filePath)
			return
		} else {
			fmt.Println("Invalid command. Usage: download_file -o <file_path> <torrent_file> <piece_index>")
			return
		}
	} else if command == "download" {
		option := os.Args[2]
		if option == "-o" && len(os.Args) == 5 {
			filePath := os.Args[3]
			fileName := os.Args[4]

			torrent, err := decodeFile(fileName)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			peersList, err := protocol.GetPeers(torrent)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			// just use the first peer, since there's no specification
			conn := protocol.EstablishTCPConnection(peersList[0])
			defer conn.Close()

			response := protocol.SendTCPHandshake(conn, torrent)
			if (response == nil) || len(response) == 0 || (conn == nil) {
				fmt.Println("Error sending handshake")
				return
			}

			data := protocol.Download(conn, torrent)
			if (data == nil) || (len(data) == 0) {
				fmt.Println("Error downloading data")
				return
			}

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Printf("Downloaded %s to %s.\n", fileName, filePath)
			return
		} else {
			fmt.Println("Invalid command. Usage: download -o <file_path> <torrent_file>")
			return
		}
	} else if command == "download_x" {
		option := os.Args[2]
		if option == "-o" && len(os.Args) == 5 {
			filePath := os.Args[3]
			fileName := os.Args[4]

			torrent, err := decodeFile(fileName)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			peersList, err := protocol.GetPeers(torrent)
			if err != nil {
				tracerr.PrintSourceColor(err)
				return
			}

			validPeers := protocol.InitPeers(peersList, torrent)
			fmt.Printf("Downloading %s from %v peers\n", fileName, validPeers)

			// initiatilizing ctx
			ctx, cancel := context.WithCancel(context.Background())

			sigCh := make(chan os.Signal, 1)
			defer close(sigCh)

			signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)
			go func() {
				// wait until receiving the signal
				<-sigCh
				cancel()
			}()

			dl := worker.NewDownloader(validPeers, torrent.Info.Length)
			d := worker.NewDispatcher(dl, len(validPeers), 5) // maxWorkers equals valid peers for now

			// split the file into pieces
			piecesHash := bencode.SplitPiecesIntoHashes(torrent.Info.Pieces)

			// add a job for each piece
			for i, pieceHash := range piecesHash {
				job := &worker.DownloadPieceJob{
					PieceIndex: i,
					Torrent:    &torrent,
					Hash:       pieceHash,
				}
				d.Add(job)
			}

			// start the dispatcher
			d.Start(ctx)

			// wait for all the jobs to finish
			d.Wait()

			// close any open connections
			dl.CloseConnections()

			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.Write(dl.FullData)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Printf("Downloaded %s to %s.\n", fileName, filePath)
			return
		} else {
			fmt.Println("Invalid command. Usage: download_x -o <file_path> <torrent_file>")
			return
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
