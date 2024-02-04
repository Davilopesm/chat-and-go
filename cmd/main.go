package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"text/template"

	"golang.org/x/net/websocket"
)

type Server struct {
	connections map[*websocket.Conn]string
	mutex       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]string),
	}
}

func (server *Server) handleWebSocket(ws *websocket.Conn) {
	fmt.Println("Connection request received", ws.RemoteAddr())

	server.mutex.Lock()
	username := fmt.Sprintf("user%d", rand.Intn(1000))
	server.connections[ws] = username
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
		server.broadcast([]byte(string(server.connections[ws]) + ": " + string(message)))
	}
}

func (server *Server) broadcast(message []byte) {
	for ws := range server.connections {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(message); err != nil {
				fmt.Println("Error: ", string(message))
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", ServeHTML)

	http.Handle("/ws", websocket.Handler(server.handleWebSocket))
	http.ListenAndServe(":3000", nil)
}
