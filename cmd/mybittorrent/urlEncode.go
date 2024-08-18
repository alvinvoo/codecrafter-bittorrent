package main

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
