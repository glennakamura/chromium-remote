package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/websocket"
)

const (
	chromiumURL = "ws://chromium:8080/chromium"
	originURL   = "http://localhost"
)

func sigPipeIgnore() {
	sigPipe := make(chan os.Signal, 4)
	signal.Notify(sigPipe, syscall.SIGPIPE)
	for {
		<-sigPipe
	}
}

func main() {
	// Ignore SIGPIPE to prevent program exit when writing to a broken pipe
	go sigPipeIgnore()
	ws, err := websocket.Dial(chromiumURL, "", originURL)
	if err != nil {
		log.Fatal(err)
	}
	if err = websocket.JSON.Send(ws, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 4096)
	for {
		num, err := ws.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write(buf[:num])
	}
}
