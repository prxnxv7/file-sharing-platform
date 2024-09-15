package handlers

import (
	"net/http"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	uploadCompleteChan = make(chan bool)
)

func NotifyUploadComplete(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		uploadComplete := <-uploadCompleteChan 

		if err := conn.WriteMessage(websocket.TextMessage, []byte("Upload completed!")); err != nil {
			return
		}

		if !uploadComplete {
			break
		}
	}
}
