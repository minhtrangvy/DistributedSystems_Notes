package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	// connect to server
	conn, err := net.Dial("tcp", ":15440")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to port 15440: %v\n", err)
		os.Exit(1)
	}

	for {
		// get message from stdin
		fmt.Println("Say something: ")
		message, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}

		// send message to server
		fmt.Fprintf(conn, "%v\n", message)

		// receive message from server
		message, err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from server: %v\n", err)
			os.Exit(1)
		}

		// print server message to Stdout
		fmt.Fprintf(os.Stdout, "%v\n", message)
	}
}
