package main

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func echoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	http.Handle("/", websocket.Handler(echoServer))

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("http.ListenAndServe: %v", err)
	}
}
