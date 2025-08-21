package main

import (
	"database/sql"
	// "fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Shredder42/space_invasion/server/internal/database"
	"github.com/Shredder42/space_invasion/shared"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	db        *database.Queries
	platform  string
	jwtSecret string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GameServer struct {
	apiConfig
	players        map[string]*shared.Player
	connections    map[string]*websocket.Conn
	connToPlayer   map[*websocket.Conn]string
	bullets        map[int]*shared.Bullet
	enemies        [][]*shared.Enemy
	running        bool
	mu             sync.Mutex
	startCondition *sync.Cond
}

func (gs *GameServer) addNewPlayer(conn *websocket.Conn, userName string) string {
	playerLocation := 300.0
	if len(gs.players) == 1 {
		playerLocation = 600.0
	}

	newPlayer := &shared.Player{
		ID: userName,
		X:  playerLocation,
		Y:  float64(shared.ScreenHeight) - 512.0*shared.ScalePlayer,
	}

	gs.players[userName] = newPlayer
	gs.connections[userName] = conn
	gs.connToPlayer[conn] = userName

	message := shared.ServerMessage{
		Type:     "player_id",
		PlayerID: userName,
	}

	conn.WriteJSON(message)

	gs.broadcastGameState()

	return ""
}

// remove player when disconnect

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

func (gs *GameServer) startGameLoop() {
	log.Println("Waiting for 2 players")

	// wait for a broadcast signal when another player connects
	gs.mu.Lock()
	for len(gs.players) < 2 {
		gs.startCondition.Wait()
	}
	gs.mu.Unlock()

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
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatalf("PLATFORM must be set")
	}

	jwtSecret := os.Getenv("SECRET")
	if jwtSecret == "" {
		log.Fatalf("SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		db:        dbQueries,
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	gameServer := &GameServer{
		apiConfig:      apiCfg,
		players:        map[string]*shared.Player{},
		connections:    map[string]*websocket.Conn{},
		connToPlayer:   map[*websocket.Conn]string{},
		bullets:        map[int]*shared.Bullet{},
		enemies:        createFleet(),
		startCondition: sync.NewCond(&sync.Mutex{}),
	}

	gameServer.startCondition = sync.NewCond(&gameServer.mu)

	const port = "8080"

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", gameServer.handlerReadiness)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUsers)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	// web socket for game
	mux.HandleFunc("/ws", gameServer.handlerWebSocket)

	go gameServer.startGameLoop()

	log.Printf("Server started on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
