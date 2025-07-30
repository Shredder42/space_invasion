package shared

import (
	"time"
	// "github.com/hajimehoshi/ebiten/v2"
	// "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	ScreenWidth  = 900
	ScreenHeight = 600
	ScalePlayer  = 1.0 / 16.0
	ScaleEnemy   = 4.0
	Cooldown     = 300 * time.Millisecond
)

type Player struct {
	ID        string    `json:"id"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	ShootTime time.Time `json:"shootTime"`
}

type PlayerAction struct {
	Type      string `json:"type"`
	Direction string `json:"direction"`
}

type GameState struct {
	Players []Player `json:"players"`
}

type ServerMessage struct {
	Type      string     `json:"type"`
	PlayerID  string     `json:"id,omitempty"`
	GameState *GameState `json:"game_state,omitempty"`
}
