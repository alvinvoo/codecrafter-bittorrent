package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decodeFile(fileName string) (TorrentMetadata, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return TorrentMetadata{}, tracerr.Wrap(err)
	}

	metadataMap, rest, err := decodeBencode(string(content))
	if err != nil {
		fmt.Println("decodeBencode error:", err)
		return TorrentMetadata{}, tracerr.Wrap(err)
	}

	if rest != "" {
		fmt.Println("Rest is not empty. Invalid syntax")
		return TorrentMetadata{}, tracerr.Wrap(err)
	}

	// Type assertion to convert interface{} to map[string]interface{}
	decodedMap, ok := metadataMap.(map[string]interface{})
	if !ok {
		fmt.Println("Failed to type assert metaMap to map[string]interface{}")
		return TorrentMetadata{}, tracerr.Wrap(err)
	}

	// Convert the decoded map to the TorrentMetadata struct
	var torrent TorrentMetadata
	if announce, ok := decodedMap["announce"].(string); ok {
		torrent.Announce = announce
	}

	if infoMap, ok := decodedMap["info"].(map[string]interface{}); ok {
		var info InfoDict
		if length, ok := infoMap["length"].(int); ok {
			info.Length = length
		}
		if name, ok := infoMap["name"].(string); ok {
			info.Name = name
		}
		if pieceLength, ok := infoMap["piece length"].(int); ok {
			info.PieceLength = pieceLength
		}
		if pieces, ok := infoMap["pieces"].(string); ok {
			info.Pieces = []byte(pieces) // pieces are non-UTF-8 bytes
		}
		torrent.Info = info
	}

	return torrent, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, rest, err := decodeBencode(bencodedValue)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		if rest != "" {
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
		fmt.Printf("Info Hash: %s\n", calculateInfoHash(torrent))
		fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
		fmt.Printf("Piece Hashes:\n")
		pieces := splitPiecesIntoHashes(torrent.Info.Pieces)
		for p := range pieces {
			fmt.Printf("%s\n", pieces[p])
		}
	} else if command == "peers" {
		fileName := os.Args[2]

		torrent, err := decodeFile(fileName)
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		infoHash := calculateInfoHash(torrent)

		url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&uploaded=0&downloaded=0&left=92063&compact=1",
			torrent.Announce, urlEncodeWithConversion(infoHash), "ALVINVOOALVINVOO1234", 6881)

		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error getting peers:", err)
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
		}

		fmt.Println(body)

		// TODO: convert bencodeString finally to process []byte
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
