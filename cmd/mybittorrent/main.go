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

		// Convert the map to JSON
		jsonData, err := json.Marshal(metadataMap)
		if err != nil {
			fmt.Println("Error marshalling to JSON:", err)
			return
		}

		// Unmarshal JSON into our struct
		var metadata TorrentMetadata
		err = json.Unmarshal(jsonData, &metadata)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		// Now you can use the struct
		fmt.Printf("Tracker URL: %s\n", metadata.Announce)
		fmt.Printf("Length: %d\n", metadata.Info.Length)
		fmt.Printf("Info Hash: %s\n", calculateInfoHash(metadata))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
