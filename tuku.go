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
	file   = flag.String("file", "", "A file to tail")
	port   = flag.Int("port", 0, "A port in which the server will bind to")
	filter = flag.String("filter", "", "A pattern to match")

	clients sockets
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
	go listenAndBroadcast(clients, messageChan)

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

// listenAndBroadcast range over through messageChan and broadcast message
// to all the connected clients. Also prints to standard out
func listenAndBroadcast(clients sockets, messageChan chan string) {
	for msg := range messageChan {
		fn := pop(*file, "/")
		log.Printf("[ %s ] %s", fn, msg)

		if regexp.MustCompile(`(?i)` + *filter).MatchString(msg) {
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
	cmd := exec.Command("tail", "-f", file)

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

func socketHandler(conn *websocket.Conn) {
	id := uuid.NewV4().String()
	clients[id] = conn

	_ = websocket.Message.Send(conn, "Welcome!")
	log.Printf("client %s connected\n", id)

	var msg string
	for websocket.Message.Receive(conn, &msg) == nil {
	}

	// If this line gets evaluated, that means the client has been disconnected.
	// So let's remove it.
	delete(clients, id)
	log.Printf("client %s disconnected (%d total clients)\n", id, len(clients))
}
