package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

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

		content, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		metadataMap, rest, err := decodeBencode(string(content))
		if err != nil {
			tracerr.PrintSourceColor(err)
			return
		}

		if rest != "" {
			fmt.Println("Rest is not empty. Invalid syntax")
			return
		}

		// Type assertion to convert interface{} to map[string]interface{}
		decodedMap, ok := metadataMap.(map[string]interface{})
		if !ok {
			fmt.Println("Failed to type assert metaMap to map[string]interface{}")
			return
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

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
