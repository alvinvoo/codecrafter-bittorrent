package protocol

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Unreserved characters in URL encoding
func isUnreserved(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') ||
		b == '-' || b == '_' ||
		b == '.' || b == '~'
}

func urlEncodeWithConversion(input string) string {
	// Decode the hex string into bytes
	data, err := hex.DecodeString(input)
	if err != nil {
		fmt.Println("Error decoding hex:", err)
		return ""
	}

	var result strings.Builder
	for _, b := range data {
		if isUnreserved(b) {
			// Convert unreserved characters to their symbol
			result.WriteByte(b)
		} else {
			// Percent encode the rest
			result.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}

	return result.String()
}

func decodePeer(peer []byte) (string, error) {
	if len(peer) != 6 {
		return "", fmt.Errorf("invalid peer length")
	}

	ip := fmt.Sprintf("%d.%d.%d.%d", peer[0], peer[1], peer[2], peer[3])

	// multi-byte integers are split across mulitple bytes in big-endian order
	// shift first 8 bits `201` to the left, placing it in the higer order byte position of a 16 bit integer
	// alternatively, can also use encoding/binary's `binary.BigEndian.Uint16(byteSlice)`
	port := int(peer[4])<<8 + int(peer[5])

	return fmt.Sprintf("%s:%d", ip, port), nil
}

func decodePeers(peers []byte) ([]string, error) {
	if (len(peers) % 6) != 0 {
		return nil, fmt.Errorf("invalid peers length")
	}

	peersList := make([]string, 0, len(peers)/6)
	for i := 0; i < len(peers); i += 6 {
		peer, err := decodePeer(peers[i : i+6])
		if err != nil {
			return nil, err
		}

		peersList = append(peersList, peer)
	}

	return peersList, nil
}
