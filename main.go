package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
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

	// parse body (everything after the empty line)
	body := ""
	if headerBodySplit > 0 && headerBodySplit < len(lines)-1 {
		body = strings.Join(lines[headerBodySplit+1:], "\r\n")
	}

	return &Request{
		Method:      method,
		Path:        path,
		Headers:     headers,
		Body:        body,
		QueryParams: queryParams,
	}
}

// matchRoute tries to match a request to a route
func (s *Server) matchRoute(req *Request) (func(request *Request, response *Response), bool) {
	for _, route := range s.routes {
		// Check if method and path match
		if route.method == req.Method && route.path == req.Path {
			return route.handler, true
		}
	}

	return nil, false
}

// formatResponse formats a Response struct into an HTTP response string
func formatResponse(resp *Response) string {
	// Start with status line
	statusText := "OK"
	if resp.StatusCode != 200 {
		statusText = "Not Found"
	} else if resp.StatusCode == 500 {
		statusText = "Internal Server Error"
	}

	result := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, statusText)

	// Add Date header if not present
	if _, ok := resp.Headers["Date"]; !ok {
		resp.Headers["Date"] = time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	}

	// Add Content-Length header if not present
	if _, ok := resp.Headers["Content-Length"]; !ok {
		resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))
	}

	// Add headers
	for key, value := range resp.Headers {
		result += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	// add empty line and body
	result += "\r\n" + resp.Body

	return result
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Print connection info
	fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

	// Read request data
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	requestString := string(buffer[:n])
	fmt.Printf("Received request:\n%s\n", requestString)

	// Parse the HTTP request
	request := parseHttpRequest(requestString)

	// Create a default response
	response := &Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/plain",
			"Server":       "GoCustomHTTP/1.0",
		},
		Body: "",
	}

	// try to match a route
	if handler, found := s.matchRoute(request); found {
		handler(request, response)
	} else {
		// No route found, send 404 response
		response.StatusCode = 404
		response.Body = "404 Not Found: " + request.Path
	}

	// Format and send the response
	responseString := formatResponse(response)
	conn.Write([]byte(responseString))
}

// Start the server on the specified address
func (s *Server) Start(address string) error {
	// create a tcp listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server started on %s\n", address)

	// accept connections in a loop
	for {
		fmt.Println("Waiting for connection...")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func main() {
	// create a new server
	server := NewServer()

	// add some routes
	server.AddRoute("GET", "/", func(req *Request, resp *Response) {
		resp.Headers["Content-Type"] = "text/html"
		resp.Body = `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Custom HTTP Server</title>
		</head>
		<body>
			<h1>Welcome to our custom HTTP server!</h1>
			<p>This server is built from scratch in Go without using the standard HTTP library.</p>
			<ul>
				<li><a href="/hello">Hello Page</a></li>
				<li><a href="/time">Current Time</a></li>
				<li><a href="/echo?message=Hello">Echo Page</a></li>
			</ul>
		</body>
		</html>
		`
	})

	server.AddRoute("GET", "/hello", func(req *Request, resp *Response) {
		resp.Headers["Content-Type"] = "text/plain"
		resp.Body = "Hello, World!"
	})

	server.AddRoute("GET", "/time", func(req *Request, resp *Response) {
		resp.Headers["Content-Type"] = "text/plain"
		resp.Body = "Current time: " + time.Now().Format("2006-01-02 15:04:05")
	})

	server.AddRoute("GET", "/echo", func(req *Request, resp *Response) {
		message, ok := req.QueryParams["message"]
		if !ok {
			message = "No message provided"
		}

		resp.Headers["Content-Type"] = "text/plain"
		resp.Body = "Echo: " + message
	})

	// start the server
	err := server.Start("localhost:8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
