package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/torrent"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/util"
)

const MY_PEER_ID = "00112233445566778899"
const BLOCK_LENGTH = 16384 // 16KiB, 2^14

func GetPeers(torrent torrent.TorrentMetadata) ([]string, error) {
	infoHash := bencode.CalculateInfoHash(torrent)

	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&uploaded=0&downloaded=0&left=92063&compact=1",
		torrent.Announce, urlEncodeWithConversion(infoHash), "00112233445566778899", 6881)

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

	var peersResp PeerResponse
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

	peersList, err := decodePeers(peersResp.Peers)
	if err != nil {
		return nil, err
	}

	return peersList, nil
}

func DestructureHandshakeResponse(response []byte) Handshake {
	return Handshake{
		length:   response[0],
		protocol: string(response[1:20]),
		resv:     [8]byte{},
		info:     response[28:48],
		PeerId:   response[48:68],
	}
}

// net.Conn is an interface
// net.TCPConn is a struct, that implements net.Conn
func EstablishTCPConnection(peerIpPort string) *net.TCPConn {
	// Establish a TCP connection
	conn, err := net.Dial("tcp", peerIpPort)
	if err != nil {
		fmt.Println("Error connecting:", err)
	}

	// Type assert the net.Conn to *net.TCPConn
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		fmt.Println("Error: Not a TCP connection")
		return nil
	}

	return tcpConn
}

func SendTCPHandshake(conn *net.TCPConn, metadata torrent.TorrentMetadata) []byte {
	handshakeMessage := Handshake{
		length:   byte(19),
		protocol: "BitTorrent protocol",
		resv:     [8]byte{},
		info:     metadata.Info.Hash(),
		PeerId:   []byte(MY_PEER_ID),
	}.encode()

	a, err := conn.Write(handshakeMessage)
	util.DebugLog("handshake message sent length: ", a)
	if err != nil {
		fmt.Println("Error sending handshake:", err)
	}

	response := make([]byte, 68)
	// Read the handshake response from the server
	n, err := conn.Read(response)
	util.DebugLog("handshake response received length: ", n)
	if err != nil {
		fmt.Println("Error receiving handshake response:", err)
	}

	// response will contain the entire protocol message
	// peer id is the last 20 bytes
	handshakeResponse := response[:n]
	return handshakeResponse
}

func getMsgFromConn(conn net.Conn) (byte, []byte) {
	buffer := make([]byte, 4)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message:", err)
	}
	messageLength := binary.BigEndian.Uint32(buffer)
	util.DebugLog("tcp message length to receive: ", messageLength)

	if messageLength == 0 {
		fmt.Println("Connection closed by peer")
	}
	buffer = make([]byte, messageLength)
	_, err = io.ReadFull(conn, buffer) // conn.Read doesnt work here; it might read lesser bytes than expected
	// checking the content after reading with conn.Read shows tonnes of 0 padding at the end, means data is not being read fully
	if err != nil {
		fmt.Println("Error reading message:", err)
	}

	id := buffer[0]
	content := buffer[1:]
	return id, content
}

func InitPeers(peersList []string, torrent torrent.TorrentMetadata) []Peer {
	var peers []Peer
	for _, peer := range peersList {
		conn := EstablishTCPConnection(peer)

		response := SendTCPHandshake(conn, torrent)

		if (response != nil) && len(response) != -1 && (conn != nil) {
			peers = append(peers, Peer{
				Conn: conn, // needs to be closed later on
				Id:   fmt.Sprintf("%x", DestructureHandshakeResponse(response).PeerId),
			})
		}
	}

	return peers
}

func downloadInit(conn net.Conn) net.Conn {
	// wait for the first message from the peer
	id, _ := getMsgFromConn(conn)
	if id != 5 {
		fmt.Println("Expected bitfield message, got something else")
		return nil
	}

	// show interested
	interestedMessage := PeerMessage{
		Length: 5,
		Id:     byte(2),
	}.encode()
	_, err := conn.Write(interestedMessage)
	if err != nil {
		fmt.Println("Error sending interested message:", err)
		return nil
	}

	util.DebugLog("Sent interested message")

	id, _ = getMsgFromConn(conn)
	if id != 1 {
		fmt.Println("Expected unchoke message, got something else")
		return nil
	}

	return conn
}

func requestPiece(conn net.Conn, torrentMetadata *torrent.TorrentMetadata, pieceIndex int) []byte {
	var pieceLengthToRetrive int
	length := torrentMetadata.Info.Length
	pieceLength := torrentMetadata.Info.PieceLength
	if (pieceIndex == length/pieceLength) && (length%pieceLength != 0) {
		// last piece
		pieceLengthToRetrive = length % pieceLength
	} else {
		pieceLengthToRetrive = pieceLength
	}

	var data []byte
	for begin := 0; begin < pieceLengthToRetrive; begin += BLOCK_LENGTH {
		actualBlockLength := BLOCK_LENGTH
		if begin+BLOCK_LENGTH > pieceLengthToRetrive {
			actualBlockLength = pieceLengthToRetrive - begin
		}

		peerMessage := PeerMessage{
			Length: 13, // 1 + 4 + 4 + 4
			Id:     6,
			index:  uint32(pieceIndex),
			begin:  uint32(begin),
			block:  uint32(actualBlockLength),
		}

		// send request message
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, peerMessage)
		_, err := conn.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error sending request message:", err)
			return nil
		}

		// read the piece
		id, content := getMsgFromConn(conn)
		if id != 7 {
			fmt.Println("Expected piece message, got ", id)
			return nil
		}

		// write the piece to the file
		data = append(data, content[8:]...) // first 8 bytes are index and begin; probably better to check
	}

	return data
}

func DownloadPiece(conn net.Conn, torrent torrent.TorrentMetadata, pieceIndex int) []byte {
	// first check pieceIndex validity
	if pieceIndex > (torrent.Info.Length/torrent.Info.PieceLength) || (pieceIndex < 0) {
		fmt.Println("Invalid piece index")
		return nil
	}

	conn = downloadInit(conn)

	return requestPiece(conn, &torrent, pieceIndex)
}

func Download(conn net.Conn, torrent torrent.TorrentMetadata) []byte {
	piecesHash := bencode.SplitPiecesIntoHashes(torrent.Info.Pieces)

	data := make([]byte, 0)

	conn = downloadInit(conn)
	for i := 0; i < len(piecesHash); i++ {
		piece := requestPiece(conn, &torrent, i)
		if util.GenerateSHA1Checksum(piece) != piecesHash[i] {
			fmt.Printf("Sha1 Checksum for Piece %d does not match\n", i)
			return nil
		}

		data = append(data, piece...)
	}

	return data
}
