package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	bullets      []*shared.Bullet
	connections  map[string]*websocket.Conn
	connToPlayer map[*websocket.Conn]string
}

func (gs *GameServer) addNewPlayer(conn *websocket.Conn) string {
	playerID := fmt.Sprintf("player_%d", len(gs.players)+1)
	playerLocation := 300.0
	if playerID == "player_2" {
		playerLocation = 600.0
	}

	newPlayer := &shared.Player{
		ID: playerID,
		// X:  float64(shared.ScreenWidth)/2.0 - 512.0*shared.ScalePlayer/2.0,
		X: playerLocation,
		Y: float64(shared.ScreenHeight) - 512.0*shared.ScalePlayer,
	}

	gs.players[playerID] = newPlayer
	gs.connections[playerID] = conn
	gs.connToPlayer[conn] = playerID

	message := shared.ServerMessage{
		Type:     "player_id",
		PlayerID: playerID,
	}

	conn.WriteJSON(message)

	gs.broadcastGameState()

	return playerID
}

func (gs *GameServer) shoot(playerID string) {
	if gs.players[playerID].ShootTime.Before(time.Now().Add(-shared.Cooldown)) {
		gs.players[playerID].ShootTime = time.Now()
		newBullet := &shared.Bullet{
			X: int(gs.players[playerID].X + 16.0), // probably should make these dynamic if possible
			Y: int(gs.players[playerID].Y - 6.0),  // probably should make these dynamic if possible
		}

		gs.bullets = append(gs.bullets, newBullet)

	}
}

// need to remove player when disconnect

func (gs *GameServer) broadcastGameState() {
	players := make([]shared.Player, 0, len(gs.players))
	for _, player := range gs.players {
		players = append(players, *player)
	}

	bullets := make([]shared.Bullet, 0, len(gs.bullets))
	for _, bullet := range gs.bullets {
		bullets = append(bullets, *bullet)
	}

	message := shared.ServerMessage{
		Type: "game_state",
		GameState: &shared.GameState{
			Players: players,
			Bullets: bullets,
		},
	}

	for _, conn := range gs.connections {
		conn.WriteJSON(message)
	}

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

		// log.Printf("action received: %+v", action)
		if action.Type == "move" {
			gs.players[action.ID].MovePlayer(action.Direction)
		}
		if action.Type == "shoot" {
			log.Printf("action received: %+v", action)
			gs.shoot(action.ID)
		}
		gs.broadcastGameState()
		// && action.Direction == "left" {
		// 	gs.p
		// set playerX to -4 or whatever
		// broadcast the game state
		// }
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
