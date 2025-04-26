package main

import (
	"log"
	"net"
)

func basicHandler(rq *Request, conn net.Conn) {
	respBody := "Sucessfully Received Data"
	headers := map[string]string{
		"Content-Length": "",
		"Content-Type":   "text/plain",
	}

	response := NewResponse(200, rq.Protocol, respBody, headers)
	_, err := conn.Write([]byte(response.Build()))
	if err != nil {
		return
	}

	return

}

func main() {
	serve := NewServer(":3000")

	serve.mux.AddRoute("POST", "/", basicHandler)
	log.Fatal(serve.Start())
}
