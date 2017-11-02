# TUKU!

&nbsp;

### Description
A daemon for tailing a given file and broadcast updates to connected clients over websocket.

### Requirements
- Go 1.9 or higher

### Usage

- Do `go get -u github.com/elizar/tuku`
- then run `tuku -file <some_file.logs> [-port:8082] [-filter]`
- Connect from any ws client using: `ws://<hostname>:8082`

### But Why?

¯\\\_(ツ)\_/¯

### TODO

- [ ] Go releaser