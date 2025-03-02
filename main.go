package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	// create a tcp listener on port 8080
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Failed to create listener:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started on :8080")

	for {
		fmt.Println("Waiting for connection...")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Print connection info
	fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

	// read data from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// print the data received
	fmt.Printf("Received %d bytes\n", n)
	fmt.Println(string(buffer[:n]))

	// send a response
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 18\r\n" +
		"\r\n" +
		"Hello from server!\n"
	conn.Write([]byte(response))
}

// parseHttpRequest parses an HTTP request string into its components
func parseHttpRequest(request string) (method, path string, headers map[string]string, body string) {
	// initialize the headers map
	headers = make(map[string]string)

	// split the request into lines
	lines := strings.Split(request, "\r\n")

	// parse the request lien (first line)
	if len(lines) > 0 {
		requestLineParts := strings.Split(lines[0], " ")
		if len(requestLineParts) >= 3 {
			method = requestLineParts[0]
			path = requestLineParts[1]
		}
	}

	// find where headers end and body begins
	headerBodySplit := -1
	for i, line := range lines {
		if line == "" {
			headerBodySplit = i
			break
		}
	}

	// parse headers (skip first line which is the request line)
	for i :=1; i < headerBodySplit; i++ {
		if lines[i] == "" {
			continue
		}

		parts := strings.SplitN(lines[i], ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}
}