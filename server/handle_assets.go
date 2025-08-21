package main

import (
	"time"

	"github.com/Shredder42/space_invasion/shared"
)

func createFleet() [][]*shared.Enemy {
	availableSpaceX := shared.ScreenWidth - (2 * 12.0 * shared.ScaleEnemy)
	numberEnemiesX := int(availableSpaceX / (2 * 12.0 * shared.ScaleEnemy))

	enemyFleet := [][]*shared.Enemy{}
	for i := 0; i < 4; i++ {
		enemyRow := []*shared.Enemy{}
		xOffset := 0.0
		if i%2 == 1 {
			xOffset = 3.0
		}
		for j := 0; j < numberEnemiesX; j++ {
			enemyRow = append(enemyRow, &shared.Enemy{
				ID:           string(i) + "_" + string(j),
				X:            12.0*shared.ScaleEnemy + 2*12.0*shared.ScaleEnemy*float64(j) + xOffset,
				Y:            25.0 + 50*float64(i),
				Health:       1,
				Speed:        1.0,
				DropDistance: 15.0,
				Width:        12.0 * shared.ScaleEnemy,
				// animations:   animation,
			})
		}
		enemyFleet = append(enemyFleet, enemyRow)
	}

	return enemyFleet
}

func (gs *GameServer) shoot(playerID string) {
	if gs.players[playerID].ShootTime.Before(time.Now().Add(-shared.Cooldown)) {
		gs.players[playerID].ShootTime = time.Now()
		shared.BulletID += 1
		newBullet := &shared.Bullet{
			ID:       shared.BulletID,
			PlayerID: playerID,
			X:        gs.players[playerID].X + 16.0,
			Y:        gs.players[playerID].Y - 6.0,
		}

		gs.bullets[shared.BulletID] = newBullet

	}
}

func (gs *GameServer) updateBullets() {

	for _, bullet := range gs.bullets {
		bullet.Y -= 4.0

		if bullet.Y < -6.0 {
			gs.removeBullet(bullet)
		}
	}
}

func (gs *GameServer) removeBullet(b *shared.Bullet) {
	delete(gs.bullets, b.ID)
}

func (gs *GameServer) updateEnemies() {
	hitEdge := false
	for _, row := range gs.enemies {
		if len(row) > 0 {
			for _, enemy := range row {
				if enemy.CheckEdges() {
					hitEdge = true
					break
				}
			}
		}
		if hitEdge {
			break
		}
	}

	for _, row := range gs.enemies {
		for _, enemy := range row {
			enemy.ChangeDirection(hitEdge)
			enemy.Move()
		}
	}

}

func (gs *GameServer) removeEnemy(target *shared.Enemy) {
	for rowIdx, row := range gs.enemies {
		for colIdx, enemy := range row {
			if enemy == target {
				gs.enemies[rowIdx] = append(row[:colIdx], row[colIdx+1:]...)
				return
			}
		}
	}
}

func (gs *GameServer) detectEnemyBulletCollision(e *shared.Enemy, b *shared.Bullet) bool {
	if e.X < b.X+3.0 &&
		e.X+12.0*shared.ScaleEnemy > b.X &&
		e.Y < b.Y+6.0 &&
		e.Y+10.0*shared.ScaleEnemy-5.0 > b.Y {
		return true
	}
	return false
}

func (gs *GameServer) handleEnemyBulletCollisions() {
	bulletHits := []*shared.Bullet{}
	enemyHits := []*shared.Enemy{}

	for _, row := range gs.enemies {
		for _, enemy := range row {
			for _, bullet := range gs.bullets {
				if gs.detectEnemyBulletCollision(enemy, bullet) {
					bulletHits = append(bulletHits, bullet)
					enemyHits = append(enemyHits, enemy)
					gs.players[bullet.PlayerID].Score += enemy.Health
				}
			}
		}
	}

	for _, hitEnemy := range enemyHits {
		gs.removeEnemy(hitEnemy)
	}

	for _, bullet := range bulletHits {
		gs.removeBullet(bullet)
	}
}
