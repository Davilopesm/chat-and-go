package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"text/template"

	"golang.org/x/net/websocket"
)

type Server struct {
	connections map[*websocket.Conn]bool
	mutex       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (server *Server) handleWebSocket(ws *websocket.Conn) {
	fmt.Println("Connection request received", ws.RemoteAddr())

	server.mutex.Lock()
	server.connections[ws] = true
	server.mutex.Unlock()

	server.readLoop(ws)
}

func (server *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
				break
			}
			fmt.Println("Error reading buff:", err)
			continue
		}
		message := buf[:n]
		server.broadcast(message)
	}
}

func (server *Server) broadcast(b []byte) {
	for ws := range server.connections {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("Error: ", err)
			}
		}(ws)
	}
}

func ServeHTML(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	tmpl.Execute(w, nil)
}

func main() {
	server := NewServer()
	http.HandleFunc("/", ServeHTML)
	http.Handle("/ws", websocket.Handler(server.handleWebSocket))
	http.ListenAndServe(":3000", nil)
}
