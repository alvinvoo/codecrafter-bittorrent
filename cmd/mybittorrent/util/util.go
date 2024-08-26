package util

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"
)

// Debug logger function
func DebugLog(title string, message ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println("DEBUG:", title, message)
	}
}

func GenerateSHA1Checksum(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
