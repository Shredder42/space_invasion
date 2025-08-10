package main

// type Enemy struct {
// 	X            float64
// 	Y            float64
// 	width        float64
// 	speedInTps   int
// 	frameCounter int
// 	health       int
// 	frame        int
// 	speed        float64
// 	dropDistance float64
// 	animations   map[int]*ebiten.Image
// }

// func (e *Enemy) Animate() {
// 	e.frameCounter -= 1
// 	if e.frameCounter < 0 {
// 		e.frameCounter = e.speedInTps
// 		if e.frame == 1 {
// 			e.frame = 2
// 		} else if e.frame == 2 {
// 			e.frame = 1
// 		}
// 	}
// }

// func (e *Enemy) Move() {
// 	e.X += e.speed
// }

// func (e *Enemy) checkEdges() bool {
// 	if e.X+e.width > screenWidth || e.X < 0.0 {
// 		return true
// 	}
// 	return false
// }

// func (e *Enemy) changeDirection(hitEdge bool) {
// 	if hitEdge {
// 		e.speed *= -1.0
// 		e.Y += e.dropDistance
// 	}
// }

// func createFleet(enemy1, enemy2 map[int]*ebiten.Image) [][]*Enemy {
// 	availableSpaceX := screenWidth - (2 * 12.0 * scaleEnemy) // may want to put in dynamic enemy width
// 	numberEnemiesX := int(availableSpaceX / (2 * 12.0 * scaleEnemy))

// 	enemyFleet := [][]*Enemy{}
// 	for i := 0; i < 4; i++ {
// 		enemyRow := []*Enemy{}
// 		animation := enemy1
// 		xOffset := 0.0
// 		if i%2 == 1 {
// 			animation = enemy2
// 			xOffset = 3.0
// 		}
// 		for j := 0; j < numberEnemiesX; j++ {
// 			enemyRow = append(enemyRow, &Enemy{
// 				X:            12.0*scaleEnemy + 2*12.0*scaleEnemy*float64(j) + xOffset,
// 				Y:            25.0 + 50*float64(i),
// 				width:        12.0 * scaleEnemy,
// 				speedInTps:   20,
// 				frameCounter: 20,
// 				health:       1,
// 				frame:        1,
// 				speed:        1.0,
// 				dropDistance: 15.0,
// 				animations:   animation,
// 			})
// 		}
// 		enemyFleet = append(enemyFleet, enemyRow)
// 	}

// 	return enemyFleet
// }
