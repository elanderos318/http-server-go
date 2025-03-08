package main

import (
	"fmt"
	"net"
	"strings"
)

// Route represents a server route with a handler function
type Route struct {
	method  string
	path    string
	handler func(request *Request, response *Response)
}

// Request represents an HTTP request
type Request struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        string
	QueryParams map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

// Server represents the HTTP server
type Server struct {
	routes []Route
}

// NewServer creates a new server instance
func NewServer() *Server {
	return &Server{
		routes: []Route{},
	}
}

// AddRoute adds a new route to the server
func (s *Server) AddRoute(method, path string, handler func(request *Request, response *Response)) {
	s.routes = append(s.routes, Route{
		method:  method,
		path:    path,
		handler: handler,
	})
}

// parseQueryParams parses query parameters from a URL path
func parseQueryParams(path string) (string, map[string]string) {
	params := make(map[string]string)

	// split path and query string
	parts := strings.SplitN(path, "?", 2)
	if len(parts) < 2 {
		return path, params
	}

	// parse query parameters
	queryString := parts[1]
	for _, param := range strings.Split(queryString, "&") {
		keyValue := strings.SplitN(param, "=", 2)
		if len(keyValue) == 2 {
			params[keyValue[0]] = keyValue[1]
		}
	}

	return parts[0], params
}

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

	request := string(buffer[:n])
	fmt.Printf("Received request:\n%s\n", request)

	method, path, headers, body := parseHttpRequest(request)

	fmt.Printf("Method: %s\n", method)
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Number of headers: %d\n", len(headers))
	fmt.Printf("Body length: %d bytes\n", len(body))

	// print the data received
	fmt.Printf("Received %d bytes\n", n)
	fmt.Println(request)

	// send a response
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello from server!\n"
	conn.Write([]byte(response))
}

// parseHttpRequest parses an HTTP request string into a Request struct
func parseHttpRequest(requestString string) *Request {
	headers := make(map[string]string)
	queryParams := make(map[string]string)

	// split the request into lines
	lines := strings.Split(requestString, "\r\n")

	// parse the request line (first line
	method := ""
	path := ""
	if len(lines) > 0 {
		requestLineParts := strings.Split(lines[0], " ")
		if len(requestLineParts) >= 2 {
			method = requestLineParts[0]
			rawPath := requestLineParts[1]

			// parse query parameters
			path, queryParams = parseQueryParams(rawPath)
		}
	}
}

// parseHttpRequest parses an HTTP request string into its components
func parseHttpRequest(request string) (method, path string, headers map[string]string, body string) {
	// initialize the headers map
	headers = make(map[string]string)

	// split the request into lines
	lines := strings.Split(request, "\r\n")

	// parse the request line (first line)
	if len(lines) > 0 {
		requestLineParts := strings.Split(lines[0], " ")
		if len(requestLineParts) >= 2 {
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
	for i := 1; i < headerBodySplit; i++ {
		if lines[i] == "" {
			continue
		}

		parts := strings.SplitN(lines[i], ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	// Parse body (everything after the empty line)
	if headerBodySplit > 0 && headerBodySplit < len(lines)-1 {
		body = strings.Join(lines[headerBodySplit+1:], "\r\n")
	}

	return method, path, headers, body
}
