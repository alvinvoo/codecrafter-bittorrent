package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/ztrue/tracerr"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) { // bencodedString[0] returns a byte (which shows up as Unicode when printed)
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
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if bencodedString[0] == 'i' && bencodedString[len(bencodedString)-1] == 'e' {
		numberStr := bencodedString[1 : len(bencodedString)-1]

		number, err := strconv.Atoi(numberStr)
		if err != nil {
			return "", err
		}

		return number, nil
	} else if bencodedString[0] == 'l' && bencodedString[len(bencodedString)-1] == 'e' {

		lists := bencodedString[1 : len(bencodedString)-1]

		retLists := make([]interface{}, 0)

		i := 0
		for i < len(lists) {
			if unicode.IsDigit(rune(lists[i])) {
				var firstColonIndex int

				for j := i; j < len(lists); j++ {
					if lists[j] == ':' {
						firstColonIndex = j
						break
					}
				}

				lengthStr := lists[i:firstColonIndex]

				length, err := strconv.Atoi(lengthStr)
				if err != nil {
					return "", tracerr.Wrap(err)
				}

				retLists = append(retLists, lists[firstColonIndex+1:firstColonIndex+1+length])

				i += firstColonIndex + length
			} else if lists[i] == 'i' {
				var lastIndex int

				for j := i; j < len(lists); j++ {
					if lists[j] == 'e' {
						lastIndex = j
						break
					}
				}

				numberStr := lists[i+1 : lastIndex]

				number, err := strconv.Atoi(numberStr)
				if err != nil {
					return "", err
				}

				retLists = append(retLists, number)

				i += len(numberStr) + 1
			}

			i += 1
		}

		return retLists, nil
	} else {
		return "", fmt.Errorf("only strings and integers are supported at the moment")
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			tracerr.PrintSourceColor(err)
			// fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
