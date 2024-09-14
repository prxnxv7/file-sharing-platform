package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub maintains active WebSocket connections and broadcasts messages to clients.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewHub initializes a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// RunHub starts the WebSocket hub for handling connections and broadcasting messages.
func (h *Hub) RunHub() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
			fmt.Println("Client connected")
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
				fmt.Println("Client disconnected")
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error sending message: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

// ServeWs handles WebSocket requests from clients.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	hub.register <- conn

	// Handle connection closure.
	defer func() {
		hub.unregister <- conn
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}
		hub.broadcast <- message
	}
}
