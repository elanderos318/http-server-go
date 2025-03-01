package main

import (
	"fmt"
	"net"
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
	response := "Hello from server!\n"
	conn.Write([]byte(response))
}