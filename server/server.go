package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// 1. Create server
// 2. Create Accept Loop --> This allows you to accept connections
// 3. Create a handleConn function that handles the connection
// 4. Create a ServerMux that handles the request and routes it to the proper handle func which returns the response
// 5. Create a Request struct that holds the request data
// 6. Create a Response struct that holds the response data
// 7. Create a function that builds the response
// 8. Create a function that parses the header data
// 9. Create a function that handles the request and routes it to the proper handle func which returns the response
// 10. Create a function that handles the connection and reads the request data
// 11. Create a function that handles the connection and writes the response data
// 12. Create a function that handles the connection and closes the connection

type Server struct {
	listenAddr string
	listener   net.Listener
	mux        *ServerMux
	quitchan   chan struct{}
	msgch      chan []byte
	requestch  chan Request
}

type Request struct {
	StatusCode int
	Route      string
	Method     string
	Headers    map[string]string
	Body       []byte
	Protocol   string
	SenderAddr string
	Conn       net.Conn
}

type Response struct {
	StatusCode int
	StatusText string
	Protocol   string
	Body       string
	Headers    map[string]string
}

func NewResponse(statuscode int, protocol, body string, headers map[string]string) *Response {
	st := map[int]string{
		200: "OK",
		400: "ERROR",
		500: "SERVER ERROR",
	}

	return &Response{
		StatusCode: statuscode,
		StatusText: st[statuscode],
		Protocol:   protocol,
		Body:       body,
		Headers:    headers,
	}
}

func (r *Response) Build() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s %d %s\r\n", r.Protocol, r.StatusCode, r.StatusText))

	if r.Body != "" && r.Headers["Content-Length"] == "" {
		r.Headers["Content-Length"] = fmt.Sprintf("%d", len(r.Body))
	}
	// Headers
	for key, value := range r.Headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Empty line to separate headers from body
	builder.WriteString("\r\n")

	// Body (if any)
	if r.Body != "" {
		builder.WriteString(r.Body)
	}

	return builder.String()

}

func NewServer(addr string) *Server {
	return &Server{
		listenAddr: addr,
		mux:        NewServerMux(),
		quitchan:   make(chan struct{}),
		msgch:      make(chan []byte, 10),
	}
}

func NewRequest() *Request {
	return &Request{Headers: make(map[string]string)}
}

func (sv *Server) Start() error {
	// This is a listener Loop that listens for connections
	ln, err := net.Listen("tcp", sv.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	sv.listener = ln

	go sv.acceptLoop()

	<-sv.quitchan

	return nil
}

func (sv *Server) acceptLoop() {
	for {
		conn, err := sv.listener.Accept()
		if err != nil {
			fmt.Println("Accept Err: ", err)
			continue
		}
		go sv.handleConn(conn)
	}
}

func parseHeader(data string) (key, value string) {
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			key := data[:i]
			value := strings.TrimSpace(data[i+1 : len(data)-1])
			return key, value
		}
	}
	return "", ""
}

func (sv *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	req := NewRequest()
	req.SenderAddr = conn.RemoteAddr().String()

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Read Error:", err)
		return
	}

	parts := strings.Split(line, " ")
	req.Method = strings.TrimSpace(parts[0])
	req.Route = strings.TrimSpace(parts[1])
	req.Protocol = strings.TrimSpace(parts[2])

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Read err %v\n", err)
			return
		}
		if strings.TrimSpace(line) == "" {
			break
		} else {
			key, val := parseHeader(line)
			req.Headers[key] = val
			continue
		}
	}

	//Write the data to a requests channel that is received by the server multiplexer

	if contentLength, err := strconv.Atoi(req.Headers["Content-Length"]); err == nil && contentLength > 0 {
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			fmt.Printf("Error Reading body: %v\n", err)
			return
		}
		req.Body = body
	}

	go sv.mux.route(req, conn)

	fmt.Println("Body:", req.Body)
	fmt.Println("Headers:", req.Headers)
	fmt.Println("Route:", req.Route)
	fmt.Println("Addr:", req.SenderAddr)

}

// Define Server Multiplexer that receives requests and handles
// Server Mux needs to handle the request and route it to the proper handle func which returns the response

type ServerMux struct {
	// Define the map that holds the func that handles the response
	// Keys are a string that combines <method><route>
	Routes map[string]func(request *Request, conn net.Conn)
}

func NewServerMux() *ServerMux {
	return &ServerMux{
		Routes: make(map[string]func(request *Request, conn net.Conn)),
	}
}

func (mux *ServerMux) AddRoute(method, route string, handler func(request *Request, conn net.Conn)) {
	key := method + route
	mux.Routes[key] = handler
}

func (mux *ServerMux) route(rq *Request, conn net.Conn) error {
	key := rq.Method + rq.Route
	handler := mux.Routes[key]

	handler(rq, conn)
	return nil
}
