package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	uuid "github.com/satori/go.uuid"

	"golang.org/x/net/websocket"
)

type sockets map[string]*websocket.Conn

var (
	file         = flag.String("file", "", "A file to tail")
	filter       = flag.String("filter", "", "A pattern to match")
	port         = flag.Int("port", 0, "A port in which the server will bind to")
	itemsToCache = flag.Int("items", 69, "Total number of messages to cache")

	clients sockets
	cache   []string
)

func main() {
	flag.Parse()

	// File
	if *file == "" {
		fmt.Printf("Error: file is required\n")
		os.Exit(2)
	}

	// Check if file exists
	_, err := os.Stat(*file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

	// Ports
	// defaults to 8082
	PORT := strconv.Itoa(*port)
	if !regexp.MustCompile(`[0-9]{2,}`).MatchString(PORT) {
		PORT = "8082"
	}

	// Init socket clients
	clients = make(sockets)

	errChan := make(chan error)
	messageChan := make(chan string)

	go tailFile(*file, errChan, messageChan)
	go listenBroadcastAndCache(clients, cacher, messageChan)

	if err := <-errChan; err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

	// TUKU!
	fmt.Printf("\n    TUKU!\n\n")

	http.Handle("/ws", websocket.Handler(socketHandler))
	err = http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}

// cacher add (n) number of messages to cache
func cacher(msg string) {
	items := cache
	if len(items) == *itemsToCache {
		items = cache[1:*itemsToCache]
	}
	cache = append(items, msg)
}

// listenBroadcastAndCache broadcasts message to connected clients
// and cache at least 20 items. Also prints to standard out
func listenBroadcastAndCache(
	clients sockets,
	cacher func(msg string),
	messageChan chan string,
) {
	for msg := range messageChan {
		fn := pop(*file, "/")
		log.Printf("[ %s ] %s", fn, msg)

		if regexp.MustCompile(`(?i)` + *filter).MatchString(msg) {
			cacher(msg)
			for _, c := range clients {
				_ = websocket.Message.Send(c, msg)
			}
		}
	}
}

// pop splits the strings with the given separator and returns the last
// item.
func pop(s, sp string) string {
	ls := strings.Split(s, sp)
	return ls[len(ls)-1]
}

// tailFile executes a `tail` command with a given file and
// broadcast changes to messageChan
func tailFile(file string, errChan chan error, messageChan chan string) {
	cmd := exec.Command("tail", "-F", file)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errChan <- errors.Wrap(err, "stdout")
		return
	}

	err = cmd.Start()
	if err != nil {
		errChan <- errors.Wrap(err, "could not start")
		return
	}

	errChan <- nil

	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		messageChan <- sc.Text()
	}
}

// socketHandler handles client connection
func socketHandler(conn *websocket.Conn) {
	id := uuid.NewV4().String()
	clients[id] = conn

	log.Printf("client %s connected\n", id)

	for _, m := range cache {
		_ = websocket.Message.Send(conn, m)
	}

	var msg string
	for websocket.Message.Receive(conn, &msg) == nil {
	}

	// If this line gets evaluated, that means the client has been disconnected.
	// So let's remove it.
	delete(clients, id)
	log.Printf("client %s disconnected (%d total clients)\n", id, len(clients))
}
