package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/websocket"
)

const (
	chromiumURL = "ws://chromium:8080/chromium"
	originURL   = "http://localhost"
)

func main() {
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
		fmt.Printf("%s", buf[:num])
	}
}
