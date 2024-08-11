package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decodeString(bencodedString string) (string, string, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", "", tracerr.Wrap(err)
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], bencodedString[firstColonIndex+1+length:], nil
}

func decodeInteger(bencodedString string) (int, string, error) {
	lastIndex := len(bencodedString) - 1
	for i := 1; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			lastIndex = i
			break
		}
	}

	numberStr := bencodedString[1:lastIndex]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return -1, "", err
	}

	return number, bencodedString[lastIndex+1:], nil
}

func decodeList(bencodedString string) ([]interface{}, string, error) {
	DebugLog("decodeList bencodedString: ", bencodedString)
	lastIndex := len(bencodedString) - 1
	startPoint := 1
	for i := 1; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			j := i - 1
			for j > 1 {
				if !unicode.IsDigit(rune(bencodedString[j])) {
					if i == j {
						return []interface{}{}, "", fmt.Errorf("invalid string syntax")
					}
					break
				}
				j--
			}

			length, err := strconv.Atoi(bencodedString[j+1 : i])

			if err != nil {
				return []interface{}{}, "", tracerr.Wrap(err)
			}

			i += length
			continue
		}

		if bencodedString[i] == 'e' {
			startPoint -= 1
		} else if bencodedString[i] == 'l' || bencodedString[i] == 'i' {
			startPoint += 1
		}

		if startPoint == 0 {
			lastIndex = i
			break
		}
	}

	DebugLog("lastIndex: ", lastIndex)
	DebugLog("startPoint: ", startPoint)

	lists := bencodedString[1:lastIndex]
	rest := bencodedString[lastIndex+1:]

	DebugLog("lists: ", lists)
	DebugLog("rest: ", rest)

	retLists := make([]interface{}, 0)
	for lists != "" {
		DebugLog("inside lists: ", lists)

		a, r, err := decodeBencode(lists)

		if err != nil {
			return []interface{}{}, "", tracerr.Wrap(err)
		}
		retLists = append(retLists, a)

		lists = r
	}

	DebugLog("retLists: ", retLists)

	return retLists, rest, nil
}

func decodeBencode(bencodedString string) (interface{}, string, error) {
	if unicode.IsDigit(rune(bencodedString[0])) { // bencodedString[0] returns a byte (which shows up as Unicode when printed)
		return decodeString(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInteger(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else {
		return "", "", fmt.Errorf("invalid syntax")
	}
}

// Debug logger function
func DebugLog(title string, message interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println("DEBUG:", title, message)
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

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
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
