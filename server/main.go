package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Shredder42/space_invasion/shared"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GameServer struct {
	players      map[string]*shared.Player
	connections  map[string]*websocket.Conn
	connToPlayer map[*websocket.Conn]string
}

func (gs *GameServer) addNewPlayer(conn *websocket.Conn) string {
	playerID := fmt.Sprintf("player_%d", len(gs.players)+1)

	newPlayer := &shared.Player{
		ID: playerID,
		X:  450,
		Y:  300,
	}

	gs.players[playerID] = newPlayer
	gs.connections[playerID] = conn
	gs.connToPlayer[conn] = playerID

	message := shared.ServerMessage{
		Type:     "player_id",
		PlayerID: playerID,
	}

	conn.WriteJSON(message)

	return playerID
}

func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade to websocket failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected!")

	playerID := gs.addNewPlayer(conn)
	log.Printf("Player %s connected", playerID)

	for {
		var action shared.PlayerAction
		err := conn.ReadJSON(&action)
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		// log.Printf("Received: %s", message)
		// conn.WriteMessage(messageType, message)

		log.Printf("action received: %+v", action)
	}

}

func main() {
	gameServer := &GameServer{
		players:      map[string]*shared.Player{},
		connections:  map[string]*websocket.Conn{},
		connToPlayer: map[*websocket.Conn]string{},
	}

	http.HandleFunc("/ws", gameServer.handleWebSocket)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
