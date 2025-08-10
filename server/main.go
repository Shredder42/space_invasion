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
	connections  map[string]*websocket.Conn
	connToPlayer map[*websocket.Conn]string
	bullets      map[int]*shared.Bullet
	enemies      [][]*shared.Enemy
	running      bool
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

	enemies := make([]shared.Enemy, 0, len(gs.enemies))
	for _, row := range gs.enemies {
		for _, enemy := range row {
			enemies = append(enemies, *enemy)
		}
	}

	message := shared.ServerMessage{
		Type: "game_state",
		GameState: &shared.GameState{
			Players: players,
			Bullets: bullets,
			Enemies: enemies,
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

		if action.Type == "move" {
			gs.players[action.ID].MovePlayer(action.Direction)
		}
		if action.Type == "shoot" {
			gs.shoot(action.ID)
		}
	}

}

func (gs *GameServer) startGameLoop() {
	gs.running = true
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	log.Println("Game loop started")

	for gs.running {
		select {
		case <-ticker.C:
			gs.updateEnemies()
			gs.updateBullets()
			gs.handleEnemyBulletCollisions()
			gs.broadcastGameState()
		}
	}
}

func main() {
	gameServer := &GameServer{
		players:      map[string]*shared.Player{},
		connections:  map[string]*websocket.Conn{},
		connToPlayer: map[*websocket.Conn]string{},
		bullets:      map[int]*shared.Bullet{},
		enemies:      createFleet(),
	}

	http.HandleFunc("/ws", gameServer.handleWebSocket)

	go gameServer.startGameLoop()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
