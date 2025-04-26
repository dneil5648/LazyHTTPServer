package main

import (
	"log"
	"net"

	sv "github.com/dneil5648/LazyHTTPServer/server"
)

func basicHandler(rq *sv.Request, conn net.Conn) {
	respBody := "Sucessfully Received Data"
	headers := map[string]string{
		"Content-Length": "",
		"Content-Type":   "text/plain",
	}

	response := sv.NewResponse(200, rq.Protocol, respBody, headers)
	_, err := conn.Write([]byte(response.Build()))
	if err != nil {
		return
	}

	return

}

func main() {
	serve := sv.NewServer(":3000")

	serve.Mux.AddRoute("POST", "/", basicHandler)
	log.Fatal(serve.Start())
}
