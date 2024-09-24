GO_RUN = go run

.PHONY: magnet_parse magnet_handshake
magnet_parse:
	$(GO_RUN) cmd/mybittorrent/*.go magnet_parse "magnet:?xt=urn:btih:d69f91e6b2ae4c542468d1073a71d4ea13879a7f&dn=sample.torrent&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce"
magnet_handshake:
	$(args) $(GO_RUN) cmd/mybittorrent/*.go magnet_handshake "magnet:?xt=urn:btih:d69f91e6b2ae4c542468d1073a71d4ea13879a7f&dn=sample.torrent&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce"