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

func (gs *GameServer) shoot(playerID string) {
	if gs.players[playerID].ShootTime.Before(time.Now().Add(-shared.Cooldown)) {
		gs.players[playerID].ShootTime = time.Now()
		shared.BulletID += 1
		newBullet := &shared.Bullet{
			ID:       shared.BulletID,
			PlayerID: playerID,
			X:        gs.players[playerID].X + 16.0, // probably should make these dynamic if possible
			Y:        gs.players[playerID].Y - 6.0,  // probably should make these dynamic if possible
		}

		gs.bullets[shared.BulletID] = newBullet

	}
}

func (gs *GameServer) updateBullets() {
	// deltaTime := 1.0 / 60.0 // 60 FPS
	var bulletsToDelete []int

	for bulletID, bullet := range gs.bullets {
		bullet.Y -= 4.0 //* deltaTime

		if bullet.Y < -6.0 {
			bulletsToDelete = append(bulletsToDelete, bulletID)
		}
	}

	for _, bulletID := range bulletsToDelete {
		delete(gs.bullets, bulletID)
		log.Printf("Bullet %d removed (off screen)", bulletID)
	}

	// log.Printf("bullets in gs.bullets: %d", len(gs.bullets))
}

func (gs *GameServer) updateEnemies() {
	for _, row := range gs.enemies {
		for _, enemy := range row {
			enemy.Move()
		}
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

	enemies := make([]shared.Enemy, 0, len(gs.enemies))
	for _, row := range gs.enemies {
		// enemies = append(enemies, row)
		for _, enemy := range row {
			enemies = append(enemies, *enemy)
		}

		// 		enemies = append(enemies, *enemy)
		// 	}
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
		// gs.broadcastGameState()
		// && action.Direction == "left" {
		// 	gs.p
		// set playerX to -4 or whatever
		// broadcast the game state
		// }
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
			gs.broadcastGameState()
		}
	}
}

func createFleet() [][]*shared.Enemy {
	availableSpaceX := shared.ScreenWidth - (2 * 12.0 * shared.ScaleEnemy) // may want to put in dynamic enemy width
	numberEnemiesX := int(availableSpaceX / (2 * 12.0 * shared.ScaleEnemy))

	enemyFleet := [][]*shared.Enemy{}
	for i := 0; i < 4; i++ {
		enemyRow := []*shared.Enemy{}
		// animation := enemy1
		xOffset := 0.0
		// if i%2 == 1 {
		// 	animation = enemy2
		// 	xOffset = 3.0
		// }
		for j := 0; j < numberEnemiesX; j++ {
			enemyRow = append(enemyRow, &shared.Enemy{
				ID:           string(i) + "_" + string(j),
				X:            12.0*shared.ScaleEnemy + 2*12.0*shared.ScaleEnemy*float64(j) + xOffset,
				Y:            25.0 + 50*float64(i),
				Health:       1,
				Speed:        1.0,
				DropDistance: 15.0,
				// width:        12.0 * shared.scaleEnemy,
				// animations:   animation,
			})
		}
		enemyFleet = append(enemyFleet, enemyRow)
	}

	return enemyFleet
}

func main() {
	gameServer := &GameServer{
		players:      map[string]*shared.Player{},
		connections:  map[string]*websocket.Conn{},
		connToPlayer: map[*websocket.Conn]string{},
		bullets:      map[int]*shared.Bullet{},
		enemies:      createFleet(),
	}

	go gameServer.startGameLoop()

	http.HandleFunc("/ws", gameServer.handleWebSocket)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
