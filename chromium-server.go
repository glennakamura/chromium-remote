package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/net/websocket"
)

func hostIP() string {
	addrs, err := net.LookupHost(os.Getenv("HOSTNAME"))
	if err != nil {
		return "127.0.0.1"
	}
	return addrs[0]
}

func sigChildHandler() {
	sigChild := make(chan os.Signal, 16)
	signal.Notify(sigChild, syscall.SIGCHLD)
	for {
		<-sigChild
		var status syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &status, 0, nil)
		for err == syscall.EINTR {
			pid, err = syscall.Wait4(-1, &status, 0, nil)
		}
		if err == nil {
			log.Printf("Child %d exit status (%d)", pid, status)
		}
	}
}

func chromiumServer(ws *websocket.Conn) {
	var args []string
	err := websocket.JSON.Receive(ws, &args)
	if err != nil {
		log.Print(err)
		args = nil
	}
	args = append(args, "--no-sandbox")
	args = append(args, "--disable-gpu")
	args = append(args, "--disable-software-rasterizer")
	args = append(args, "--remote-debugging-address="+hostIP())
	args = append(args, "--remote-debugging-port=0")
	cmd := exec.Command("chromium-browser", args...)
	cmd.Stdout = ws
	cmd.Stderr = ws
	err = cmd.Start()
	if err != nil {
		log.Print(err)
		return
	}
	log.Print("Start chromium-browser " + strings.Join(args, " "))
	buf := make([]byte, 512)
	for {
		if _, err := ws.Read(buf); err != nil {
			break
		}
	}
	ws.Close()
	cmd.Process.Kill()
	cmd.Wait()
	log.Print("Kill chromium-browser " + strings.Join(args, " "))
}

func main() {
	http.Handle("/chromium", websocket.Handler(chromiumServer))
	if os.Getpid() == 1 {
		go sigChildHandler()
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
