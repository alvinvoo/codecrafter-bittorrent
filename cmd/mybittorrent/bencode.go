package main

import (
	// Uncomment this line to pass the first stage
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decodeString(bencodedString []byte) ([]byte, []byte, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(string(lengthStr))
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], bencodedString[firstColonIndex+1+length:], nil
}

func decodeInteger(bencodedString []byte) (int, []byte, error) {
	lastIndex := len(bencodedString) - 1
	for i := 1; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			lastIndex = i
			break
		}
	}

	numberStr := bencodedString[1:lastIndex]

	number, err := strconv.Atoi(string(numberStr))
	if err != nil {
		return -1, nil, err
	}

	return number, bencodedString[lastIndex+1:], nil
}

func returnLastIndex(bencodedString []byte) (int, error) {
	lastIndex := len(bencodedString) - 1
	startPoint := 1

	// if there's a chain of strings
	// previous string might have integer behind it
	lastStringIndex := 0
	// this whole for loop is looking for the end `e` of this list and set it on `lastIndex`
	for i := 1; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			// dealing with string
			j := i - 1
			for j > lastStringIndex {
				if !unicode.IsDigit(rune(bencodedString[j])) {
					if i == j {
						return -1, fmt.Errorf("invalid string syntax")
					}
					break
				}
				j--
			}

			length, err := strconv.Atoi(string(bencodedString[j+1 : i]))

			if err != nil {
				return -1, tracerr.Wrap(err)
			}

			i += length
			lastStringIndex = i
			continue
		}

		t := bencodedString[i]

		if t == 'e' {
			startPoint -= 1
		} else if t == 'l' ||
			t == 'i' ||
			t == 'd' {
			startPoint += 1
		}

		if startPoint == 0 {
			lastIndex = i
			break
		}
	}

	return lastIndex, nil
}

func decodeList(bencodedString []byte) ([]interface{}, []byte, error) {
	lastIndex, err := returnLastIndex(bencodedString)

	if (err != nil) || (lastIndex == -1) {
		return []interface{}{}, nil, fmt.Errorf("invalid list syntax")
	}

	lists := bencodedString[1:lastIndex]
	rest := bencodedString[lastIndex+1:]

	retLists := make([]interface{}, 0)
	for len(lists) > 0 {
		a, r, err := decodeBencode(lists)

		if err != nil {
			return []interface{}{}, nil, tracerr.Wrap(err)
		}
		retLists = append(retLists, a)

		lists = r
	}

	return retLists, rest, nil
}

func decodeDictionary(bencodedString []byte) (map[string]interface{}, []byte, error) {
	lastIndex, err := returnLastIndex(bencodedString)

	if (err != nil) || (lastIndex == -1) {
		return map[string]interface{}{}, nil, fmt.Errorf("invalid dictionary syntax")
	}

	dict := bencodedString[1:lastIndex]
	rest := bencodedString[lastIndex+1:]

	retDict := make(map[string]interface{})
	for len(dict) > 0 {
		// key is always string
		k, rk, err := decodeString(dict)

		if err != nil {
			return map[string]interface{}{}, nil, tracerr.Wrap(err)
		}

		// need to decode twice to get key-value pair
		v, r, err := decodeBencode(rk)

		if err != nil {
			return map[string]interface{}{}, nil, tracerr.Wrap(err)
		}
		// need one way to extract out the key and the value
		retDict[string(k)] = v

		dict = r
	}

	return retDict, rest, nil
}

func decodeBencode(bencodedString []byte) (interface{}, []byte, error) {
	if unicode.IsDigit(rune(bencodedString[0])) { // bencodedString[0] returns a byte (which shows up as Unicode when printed)
		s, r, err := decodeString(bencodedString)
		if utf8.Valid(s) {
			return string(s), r, err
		}
		return s, r, err
	} else if bencodedString[0] == 'i' {
		return decodeInteger(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else if bencodedString[0] == 'd' {
		return decodeDictionary(bencodedString)
	} else {
		return nil, nil, fmt.Errorf("invalid syntax")
	}
}

func encodeInfoDict(info InfoDict) string {
	return fmt.Sprintf("d6:lengthi%de4:name%d:%s12:piece lengthi%de6:pieces%d:%se",
		info.Length, len(info.Name), info.Name, info.PieceLength, len(info.Pieces), info.Pieces)
}

func generateSHA1Checksum(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func calculateInfoHash(metadata TorrentMetadata) string {
	encodedInfoDict := encodeInfoDict(metadata.Info)
	return generateSHA1Checksum(encodedInfoDict)
}

func splitPiecesIntoHashes(pieces []byte) []string {
	hashes := make([]string, 0)
	for i := 0; i < len(pieces); i += 20 {
		hashes = append(hashes, hex.EncodeToString(pieces[i:i+20]))
	}
	return hashes
}

// Debug logger function
func DebugLog(title string, message interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println("DEBUG:", title, message)
	}
}
