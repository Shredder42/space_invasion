package main

import (
	"image/color"
	"log"

	"github.com/Shredder42/space_invasion/shared"
	"github.com/hajimehoshi/ebiten/v2"
)

type ClientPlayer struct {
	shared.Player
	Img *ebiten.Image
}

type ClientBullet struct {
	shared.Bullet
	Img *ebiten.Image
}

type ClientEnemy struct {
	shared.Enemy
	FrameCounter   int
	Frame          int
	AnimationSpeed int
	Animations     map[int]*ebiten.Image
}

func (ce *ClientEnemy) Animate() {
	ce.FrameCounter -= 1
	if ce.FrameCounter == 0 {
		ce.FrameCounter = ce.AnimationSpeed
		if ce.Frame == 1 {
			ce.Frame = 2
		} else {
			ce.Frame = 1
		}
	}
}

func (g *Game) updateClientPlayers(gameState *shared.GameState) {
	for _, serverPlayer := range gameState.Players {
		clientPlayer, exists := g.clientPlayers[serverPlayer.ID]

		if !exists {
			clientPlayerImg := g.spaceshipImg1
			if len(g.clientPlayers) == 1 {
				clientPlayerImg = g.spaceshipImg2
			}
			clientPlayer = &ClientPlayer{
				Player: serverPlayer,
				Img:    clientPlayerImg,
			}
			g.clientPlayers[serverPlayer.ID] = clientPlayer
			log.Printf("Created ClientPlayer for %s", serverPlayer.ID)
		} else {
			clientPlayer.Player = serverPlayer
		}
	}
}

func (g *Game) updateClientBullets(gameState *shared.GameState) {
	for _, serverBullet := range gameState.Bullets {
		clientBullet, exists := g.clientBullets[serverBullet.ID]

		if !exists {
			clientBullet = &ClientBullet{
				Bullet: serverBullet,
				Img:    ebiten.NewImage(3, 6),
			}
			clientBullet.Img.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})
			g.clientBullets[clientBullet.ID] = clientBullet

		} else {
			clientBullet.Bullet = serverBullet
		}
	}

	// remove client bullets that are no longer in the server bullets
	if len(gameState.Bullets) >= 1 {
		existingBullets := map[int]struct{}{}
		for _, bullet := range gameState.Bullets {
			existingBullets[bullet.ID] = struct{}{}
		}
		for id := range g.clientBullets {
			if _, ok := existingBullets[id]; !ok {
				delete(g.clientBullets, id)
			}
		}
	} else {
		g.clientBullets = map[int]*ClientBullet{}
	}

}

func (g *Game) updateClientEnemies(gameState *shared.GameState) {
	for _, serverEnemy := range gameState.Enemies {
		clientEnemy, exists := g.clientEnemyFleet[serverEnemy.ID]

		if !exists {
			animation := g.Enemy1Imgs
			if int(serverEnemy.ID[0]) == 1 || int(serverEnemy.ID[0]) == 3 {
				animation = g.Enemy2Imgs
			}
			clientEnemy := &ClientEnemy{
				Enemy:          serverEnemy,
				FrameCounter:   20,
				Frame:          1,
				AnimationSpeed: 20,
				Animations:     animation,
			}
			g.clientEnemyFleet[serverEnemy.ID] = clientEnemy
		} else {
			clientEnemy.Enemy = serverEnemy
		}
	}

	// remove client enemies that are no longer in the server enemies
	if len(gameState.Enemies) >= 1 {
		existingEnemies := map[string]struct{}{}
		for _, enemy := range gameState.Enemies {
			existingEnemies[enemy.ID] = struct{}{}
		}
		for id := range g.clientEnemyFleet {
			if _, ok := existingEnemies[id]; !ok {
				delete(g.clientEnemyFleet, id)

			}
		}
	} else {
		g.clientEnemyFleet = map[string]*ClientEnemy{}
	}
}
