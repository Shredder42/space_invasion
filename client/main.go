package main

import (
	// "fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/Shredder42/space_invasion/shared"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 900
	screenHeight = 600
	scalePlayer  = 1.0 / 16.0
	scaleEnemy   = 4.0
	cooldown     = 300 * time.Millisecond
)

type Game struct {
	BackgroundImg          *ebiten.Image
	BackgroundBuildingsImg *ebiten.Image
	spaceshipImg1          *ebiten.Image
	spaceshipImg2          *ebiten.Image
	// clientEnemyImg1        *ebiten.Image
	Enemy1Imgs map[int]*ebiten.Image
	Enemy2Imgs map[int]*ebiten.Image
	// player                 *ClientPlayer
	enemyFleet [][]*Enemy
	// bullets    []*Bullet
	conn       *websocket.Conn
	connected  bool
	myPlayerID string

	gameState *shared.GameState

	clientPlayers    map[string]*ClientPlayer
	clientBullets    map[int]*ClientBullet
	clientEnemyFleet map[string]*ClientEnemy
}

type ClientPlayer struct {
	shared.Player
	Img *ebiten.Image
	// X         float64
	// Y         float64
	// shootTime time.Time
}

type ClientBullet struct {
	shared.Bullet
	Img *ebiten.Image
	// X   float64
	// Y   float64
}

type ClientEnemy struct {
	shared.Enemy
	// Img        *ebiten.Image
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

// func (g *Game) detectEnemyBulletCollision(e *Enemy, b *Bullet) bool {
// 	if e.X < b.X+3.0 &&
// 		e.X+12.0*scaleEnemy > b.X &&
// 		e.Y < b.Y+6.0 &&
// 		e.Y+10.0*scaleEnemy-5.0 > b.Y {
// 		// fmt.Printf("enemy: %+v\n", e)
// 		return true
// 	}
// 	return false
// 	// bullet is 3x6
// }

func (g *Game) removeEnemy(target *Enemy) {
	for rowIndex, row := range g.enemyFleet {
		for colIndex, enemy := range row {
			if enemy == target {
				g.enemyFleet[rowIndex] = append(row[:colIndex], row[colIndex+1:]...)
				return
			}
		}
	}
}

// func (g *Game) removeBullet(target *Bullet) {
// 	for index, bullet := range g.bullets {
// 		if bullet == target {
// 			g.bullets = append(g.bullets[:index], g.bullets[index+1:]...)
// 			return
// 		}
// 	}
// }

func (g *Game) connectToServer() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		return
	}

	g.conn = conn
	g.connected = true
	log.Println("Connected to server!")

	// g.conn.WriteMessage(websocket.TextMessage, []byte("Hello from client!"))
}

func (g *Game) listenForServerMessages() {
	for {
		var message shared.ServerMessage
		err := g.conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			return
		}

		switch message.Type {
		case "player_id":
			log.Printf(message.PlayerID)
			g.myPlayerID = message.PlayerID
		case "game_state":
			// log.Printf("players: %v", message.GameState.Players)
			g.updateClientPlayers(message.GameState)
			g.updateBullets(message.GameState)
			g.updateEnemies(message.GameState)
			// log.Printf("enemies frame: %v", message.GameState.Enemies[0][0].Frame)
		}
	}
}

func (g *Game) updateClientPlayers(gameState *shared.GameState) {
	for _, serverPlayer := range gameState.Players {
		clientPlayer, exists := g.clientPlayers[serverPlayer.ID]

		if !exists {
			clientPlayerImg := g.spaceshipImg1
			if serverPlayer.ID == "player_2" {
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

func (g *Game) updateBullets(gameState *shared.GameState) {
	// log.Printf("Bullets shot: %d", len(gameState.Bullets))
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
			// log.Printf("bullet y: %v", clientBullet.Y)
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

	// log.Printf("length of client bullets %d", len(g.clientBullets))
}

func (g *Game) updateEnemies(gameState *shared.GameState) {
	// log.Printf("enemy frame: %d", gameState.Enemies[0][0].Frame)
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

func (g *Game) Update() error {

	// might be best to make this fleet level (also for controlling speed and health as well)

	// bulletHits := []*Bullet{}
	// enemyHits := []*Enemy{}
	// for _, row := range g.enemyFleet {
	// 	for _, enemy := range row {
	// 		for _, bullet := range g.bullets {
	// 			if g.detectEnemyBulletCollision(enemy, bullet) {
	// 				bulletHits = append(bulletHits, bullet)
	// 				enemyHits = append(enemyHits, enemy)
	// 			}
	// 		}
	// hitEdge = enemy.checkEdges()
	// if hitEdge {
	// 	break
	// }
	// 	}
	// }

	// for _, bullet := range bulletHits {
	// 	g.removeBullet(bullet)
	// }

	// for _, hitEnemy := range enemyHits {
	// 	g.removeEnemy(hitEnemy)
	// }

	hitEdge := false
	for _, row := range g.enemyFleet {
		if len(row) > 0 {
			for _, enemy := range row {
				hitEdge = enemy.checkEdges()
				if hitEdge {
					break
				}
			}
		}
		if hitEdge {
			break
		}
	}

	// fmt.Println(hitEdge)
	// for _, row := range g.enemyFleet {
	// 	for _, enemy := range row {
	// 		enemy.Animate()
	// 		enemy.changeDirection(hitEdge)
	// 		enemy.Move()
	// 	}
	// }

	for _, enemy := range g.clientEnemyFleet {
		enemy.Animate()
	}

	// for _, bullet := range g.bullets {
	// 	if bullet.Y < -4 {
	// 		g.removeBullet(bullet)
	// 	}
	// 	bullet.Y -= 4
	// }

	// make this a moveSpeed
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		action := shared.PlayerAction{ID: g.myPlayerID, Type: "move", Direction: "left"}
		g.conn.WriteJSON(action)
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		action := shared.PlayerAction{ID: g.myPlayerID, Type: "move", Direction: "right"}
		g.conn.WriteJSON(action)
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		action := shared.PlayerAction{ID: g.myPlayerID, Type: "shoot"}
		g.conn.WriteJSON(action)
	}

	// if ebiten.IsKeyPressed(ebiten.KeySpace) {
	// 	// if g.player.shootTime.Before(time.Now().Add(-cooldown)) {
	// 	// 	g.player.shootTime = time.Now()
	// 	// 	// make a create bullet function - make sense to make a bullets package - maybe just own file?
	// 	// 	newBullet := &Bullet{
	// 	// 		Img: ebiten.NewImage(3, 6),
	// 	// 		// probably should make these dynamic if possible
	// 	// 		X: g.player.X + 16,
	// 	// 		Y: g.player.Y - 6,
	// 	// 	}
	// 		newBullet.Img.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// 		g.bullets = append(g.bullets, newBullet)
	// 		// fmt.Println(g.bullets)
	// 	}
	// }

	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	backgroundImgWidth := g.BackgroundImg.Bounds().Dx()
	backgroundImgHeight := g.BackgroundImg.Bounds().Dy()

	scaleX := float64(screenWidth) / float64(backgroundImgWidth)
	scaleY := float64(screenHeight) / float64(backgroundImgHeight)

	// fmt.Println(scaleX)

	opts := ebiten.DrawImageOptions{}

	opts.GeoM.Scale(scaleX, scaleY)

	screen.DrawImage(g.BackgroundImg, &opts)
	screen.DrawImage(g.BackgroundBuildingsImg, &opts)

	opts.GeoM.Reset()

	for _, clientPlayer := range g.clientPlayers {
		opts.GeoM.Scale(scalePlayer, scalePlayer)
		opts.GeoM.Translate(clientPlayer.X, clientPlayer.Y)
		screen.DrawImage(clientPlayer.Img, &opts)
		opts.GeoM.Reset()
	}
	// opts.GeoM.Scale(scalePlayer, scalePlayer)

	// opts.GeoM.Translate(g.player.X, g.player.Y)

	// screen.DrawImage(g.player.Img, &opts)

	// for _, row := range g.enemyFleet {
	// 	for _, enemy := range row {
	// 		opts.GeoM.Scale(scaleEnemy, scaleEnemy)
	// 		opts.GeoM.Translate(enemy.X, enemy.Y)
	// 		screen.DrawImage(enemy.animations[enemy.frame], &opts)
	// 		opts.GeoM.Reset()
	// 	}
	// }

	for _, clientEnemy := range g.clientEnemyFleet {
		opts.GeoM.Scale(shared.ScaleEnemy, shared.ScaleEnemy)
		opts.GeoM.Translate(clientEnemy.X, clientEnemy.Y)
		screen.DrawImage(clientEnemy.Animations[clientEnemy.Frame], &opts)
		opts.GeoM.Reset()
	}

	for _, clientBullet := range g.clientBullets {
		opts.GeoM.Translate(clientBullet.X, clientBullet.Y)
		screen.DrawImage(clientBullet.Img, &opts)
		opts.GeoM.Reset()
	}

	// if g.connected {
	// 	ebitenutil.DebugPrint(screen, "Connected to server!")
	// } else {
	// 	ebitenutil.DebugPrint(screen, "Not connected")
	// }

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 900, 600
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Invasion!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	backgroundImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_Background.png")
	if err != nil {
		log.Fatal(err)
	}

	backgroundBuildingsImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_BackgroundBuildings.png")
	if err != nil {
		log.Fatal(err)
	}

	playerImg1, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_white_spaceship.png")
	if err != nil {
		log.Fatal(err)
	}

	playerImg2, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_blue_spaceship.png")
	if err != nil {
		log.Fatal(err)
	}

	spritesImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders.png")
	if err != nil {
		log.Fatal(err)
	}

	enemy1 := map[int]*ebiten.Image{
		1: spritesImg.SubImage(image.Rect(2, 4, 14, 14)).(*ebiten.Image),
		2: spritesImg.SubImage(image.Rect(18, 4, 30, 14)).(*ebiten.Image),
	}

	enemy2 := map[int]*ebiten.Image{
		1: spritesImg.SubImage(image.Rect(2, 20, 14, 28)).(*ebiten.Image),
		2: spritesImg.SubImage(image.Rect(18, 20, 30, 28)).(*ebiten.Image),
	}

	// clientEnemyImg1 := spritesImg.SubImage(image.Rect(2, 4, 14, 14)).(*ebiten.Image)

	game := &Game{
		BackgroundImg:          backgroundImg,
		BackgroundBuildingsImg: backgroundBuildingsImg,
		spaceshipImg1:          playerImg1,
		spaceshipImg2:          playerImg2,
		// clientEnemyImg1:        clientEnemyImg1,
		Enemy1Imgs:       enemy1,
		Enemy2Imgs:       enemy2,
		clientPlayers:    map[string]*ClientPlayer{},
		clientBullets:    map[int]*ClientBullet{},
		clientEnemyFleet: map[string]*ClientEnemy{},
		// 	player: &ClientPlayer{
		// 		Img: playerImg,
		// 		X:   float64(screenWidth)/2.0 - 512.0*scalePlayer/2.0,
		// 		Y:   float64(screenHeight) - 512.0*scalePlayer,
		// 	},
		enemyFleet: createFleet(enemy1, enemy2),
	}
	game.connectToServer()
	go game.listenForServerMessages() // not sure if this should stay as a go routine
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}
