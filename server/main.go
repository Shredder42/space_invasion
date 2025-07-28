package main

import (
	// "fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade to websocket failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected!")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		log.Printf("Received: %s", message)
		conn.WriteMessage(messageType, message)
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
