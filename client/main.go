package main

import (
	// "fmt"
	"image"
	// "image/color"
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
	// player                 *ClientPlayer
	enemyFleet [][]*Enemy
	bullets    []*Bullet
	conn       *websocket.Conn
	connected  bool
	myPlayerID string

	gameState *shared.GameState

	clientPlayers map[string]*ClientPlayer
}

type ClientPlayer struct {
	shared.Player
	Img *ebiten.Image
	// X         float64
	// Y         float64
	// shootTime time.Time
}

type Bullet struct {
	Img *ebiten.Image
	X   float64
	Y   float64
}

func (g *Game) detectEnemyBulletCollision(e *Enemy, b *Bullet) bool {
	if e.X < b.X+3.0 &&
		e.X+12.0*scaleEnemy > b.X &&
		e.Y < b.Y+6.0 &&
		e.Y+10.0*scaleEnemy-5.0 > b.Y {
		// fmt.Printf("enemy: %+v\n", e)
		return true
	}
	return false
	// bullet is 3x6
}

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

func (g *Game) removeBullet(target *Bullet) {
	for index, bullet := range g.bullets {
		if bullet == target {
			g.bullets = append(g.bullets[:index], g.bullets[index+1:]...)
			return
		}
	}
}

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
	log.Printf("Bullets shot: %d", len(gameState.Bullets))
}

func (g *Game) Update() error {

	// might be best to make this fleet level (also for controlling speed and health as well)

	bulletHits := []*Bullet{}
	enemyHits := []*Enemy{}
	for _, row := range g.enemyFleet {
		for _, enemy := range row {
			for _, bullet := range g.bullets {
				if g.detectEnemyBulletCollision(enemy, bullet) {
					bulletHits = append(bulletHits, bullet)
					enemyHits = append(enemyHits, enemy)
				}
			}
			// hitEdge = enemy.checkEdges()
			// if hitEdge {
			// 	break
			// }
		}
	}

	for _, bullet := range bulletHits {
		g.removeBullet(bullet)
	}

	for _, hitEnemy := range enemyHits {
		g.removeEnemy(hitEnemy)
	}

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
	for _, row := range g.enemyFleet {
		for _, enemy := range row {
			enemy.Animate()
			enemy.changeDirection(hitEdge)
			enemy.Move()
		}
	}

	for _, bullet := range g.bullets {
		if bullet.Y < -4 {
			g.removeBullet(bullet)
		}
		bullet.Y -= 4
	}

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
	}
	// opts.GeoM.Scale(scalePlayer, scalePlayer)

	// opts.GeoM.Translate(g.player.X, g.player.Y)

	// screen.DrawImage(g.player.Img, &opts)

	opts.GeoM.Reset()

	for _, row := range g.enemyFleet {
		for _, enemy := range row {
			opts.GeoM.Scale(scaleEnemy, scaleEnemy)
			opts.GeoM.Translate(enemy.X, enemy.Y)
			screen.DrawImage(enemy.animations[enemy.frame], &opts)
			opts.GeoM.Reset()
		}
	}

	for _, bullet := range g.bullets {
		opts.GeoM.Translate(bullet.X, bullet.Y)
		screen.DrawImage(bullet.Img, &opts)
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

	game := &Game{
		BackgroundImg:          backgroundImg,
		BackgroundBuildingsImg: backgroundBuildingsImg,
		spaceshipImg1:          playerImg1,
		spaceshipImg2:          playerImg2,
		clientPlayers:          map[string]*ClientPlayer{},
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
