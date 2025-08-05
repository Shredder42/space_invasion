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

var playerSpeed = 4.0

type Player struct {
	ID        string    `json:"id"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	ShootTime time.Time `json:"shootTime"`
}

func (p *Player) MovePlayer(d string) {
	if d == "left" {
		p.X -= playerSpeed // this could be a player speed variable
	}
	if d == "right" {
		p.X += playerSpeed
	}
}

var BulletID = 0

type Bullet struct {
	ID       int     `json:"id"`
	PlayerID string  `json:"player_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

type PlayerAction struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Direction string `json:"direction,omitempty"`
}

type GameState struct {
	Players []Player `json:"players"`
	Bullets []Bullet `json:"bullets"`
}

type ServerMessage struct {
	Type      string     `json:"type"`
	PlayerID  string     `json:"id,omitempty"`
	GameState *GameState `json:"game_state,omitempty"`
}
