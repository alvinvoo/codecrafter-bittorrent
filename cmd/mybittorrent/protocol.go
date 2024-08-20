package main

import (
	"crypto/sha1"
	"fmt"
	"net"
)

func peerHandshake(metadata TorrentMetadata) []byte {
	protocol := "BitTorrent protocol"
	lengthOfProtocol := len(protocol)
	reservedBytes := make([]byte, 8)

	encodeInfoDict := encodeInfoDict(metadata.Info)
	h := sha1.New()
	h.Write([]byte(encodeInfoDict))
	sha1InfoHash := h.Sum(nil)
	peerId := "00112233445566778899"

	// Create a byte slice and append all variables to it
	var result []byte
	result = append(result, byte(lengthOfProtocol)) // Append lengthOfProtocol (convert to byte)
	result = append(result, []byte(protocol)...)    // Append protocol
	result = append(result, reservedBytes...)       // Append reservedBytes
	result = append(result, sha1InfoHash...)        // Append sha1InfoHash
	result = append(result, []byte(peerId)...)      // Append peerId

	return result
}

func sendTCPHandshake(peerIpPort string, metadata TorrentMetadata) []byte {
	// Establish a TCP connection
	conn, err := net.Dial("tcp", peerIpPort)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}
	defer conn.Close()

	message := peerHandshake(metadata)
	a, err := conn.Write(message)
	DebugLog("tcp message sent length: ", a)
	if err != nil {
		fmt.Println("Error sending handshake:", err)
	}

	response := make([]byte, 1024)
	// Read the handshake response from the server
	n, err := conn.Read(response)
	DebugLog("tcp message received length: ", n)
	if err != nil {
		fmt.Println("Error receiving handshake response:", err)
	}

	// response will contain the entire protocol message
	// peer id is the last 20 bytes
	r := response[:n]
	return r
}

func destructureHandshakeResponse(response []byte) string {
	// response received is actually 74 bytes, different from the 68 bytes in the protocol

	// Extract the peer id from the response
	peerId := string(response[len(response)-20:])

	return peerId
}
