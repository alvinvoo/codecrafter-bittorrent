package extension

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/protocol"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/util"
)

type Magnet struct {
	link            string
	URL             string
	InfoHash        string
	InfoHashDecoded []byte
}

func NewMagnet(link string) *Magnet {
	return &Magnet{link: link}
}

func (m *Magnet) Parse() error {
	if afterStr, found := strings.CutPrefix(m.link, "magnet:?xt=urn:btih:"); found {
		parts := strings.Split(afterStr, "&")

		urls := strings.Split(parts[2], "=")
		decodedURL, err := url.QueryUnescape(urls[1])
		if err != nil {
			return err
		}

		m.URL = decodedURL
		m.InfoHash = parts[0]
		infoHashDecoded, err := hex.DecodeString(m.InfoHash)
		if err != nil {
			return err
		}

		m.InfoHashDecoded = infoHashDecoded

		return nil
	} else {
		return fmt.Errorf("invalid magnet link")
	}
}

func (m *Magnet) GetPeers() ([]string, error) {
	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&uploaded=0&downloaded=0&left=92063&compact=1",
		m.URL, protocol.UrlEncodeWithConversion(m.InfoHash), "00112233445566778899", 6881)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting peers: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	util.DebugLog("Response body: ", body)
	peersRespMap, rest, err := bencode.DecodeBencode(body)
	if err != nil {
		return nil, err
	}

	if len(rest) != 0 {
		return nil, fmt.Errorf("Rest is not empty. Invalid syntax")
	}

	util.DebugLog("Response map", peersRespMap)

	var peersResp protocol.PeerResponse
	// Type assertion to convert interface{} to map[string]interface{}
	if decodedMap, ok := peersRespMap.(map[string]interface{}); ok {
		if complete, ok := decodedMap["complete"].(int); ok {
			peersResp.Complete = complete
		}
		if incomplete, ok := decodedMap["incomplete"].(int); ok {
			peersResp.Incomplete = incomplete
		}
		if interval, ok := decodedMap["interval"].(int); ok {
			peersResp.Interval = interval
		}
		if minInterval, ok := decodedMap["min interval"].(int); ok {
			peersResp.MinInterval = minInterval
		}
		if peers, ok := decodedMap["peers"].([]byte); ok {
			peersResp.Peers = peers
		}
	}

	peersList, err := protocol.DecodePeers(peersResp.Peers)
	if err != nil {
		return nil, err
	}

	return peersList, nil
}
