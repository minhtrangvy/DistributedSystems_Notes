package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var count int
var mu sync.Mutex

func handleConnection(conn net.Conn) {
	// Update client counter
	mu.Lock()
	count++
	id := count
	mu.Unlock()

	// create reader from connection
	reader := bufio.NewReader(conn)

	// echo as messages are being received
	for {
		// read message
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from connection: %v\n", err)
			os.Exit(1)
		}
		// add server tag to message
		fmt.Fprintf(conn, "From server (you are client #%d): %v\n", id, message)

		// check for done command
		if message == "done\n" {
			conn.Close()
			return
		}
	}

}

func main() {
	// create server
	server, err := net.Listen("tcp", ":15440")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting server up: %v\n", err)
		os.Exit(1)
	}

	for {
		// accept connections
		conn, err := server.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
