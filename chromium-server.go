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
	sigChild := make(chan os.Signal, 4)
	signal.Notify(sigChild, syscall.SIGCHLD)
	for {
		<-sigChild
		for {
			var status syscall.WaitStatus
			pid, err := syscall.Wait4(-1, &status, 0, nil)
			if err == nil {
				log.Printf("Child PID:%d exit status (%d)",
					pid, status)
			} else if err == syscall.ECHILD {
				break
			}
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
	cmd := exec.Command("chromium-browser", args...)
	cmd.Stdout = ws
	cmd.Stderr = ws
	err = cmd.Start()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("Start PID:%d chromium-browser "+strings.Join(args, " "),
		cmd.Process.Pid)
	buf := make([]byte, 512)
	for {
		if _, err := ws.Read(buf); err != nil {
			break
		}
	}
	ws.Close()
	cmd.Process.Kill()
	cmd.Wait()
	log.Printf("Kill PID:%d chromium-browser", cmd.Process.Pid)
}

func main() {
	http.Handle("/chromium", websocket.Handler(chromiumServer))
	if os.Getpid() == 1 {
		go sigChildHandler()
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
