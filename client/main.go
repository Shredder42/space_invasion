package main

import (
	// "fmt"
	"image"
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
	Enemy1Imgs             map[int]*ebiten.Image
	Enemy2Imgs             map[int]*ebiten.Image

	conn       *websocket.Conn
	connected  bool
	myPlayerID string

	// gameState *shared.GameState

	clientPlayers    map[string]*ClientPlayer
	clientBullets    map[int]*ClientBullet
	clientEnemyFleet map[string]*ClientEnemy
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
			g.updateClientPlayers(message.GameState)
			g.updateClientBullets(message.GameState)
			g.updateClientEnemies(message.GameState)
		}
	}
}

func (g *Game) Update() error {

	for _, enemy := range g.clientEnemyFleet {
		enemy.Animate()
	}

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

	return nil

}

func (g *Game) Draw(screen *ebiten.Image) {

	backgroundImgWidth := g.BackgroundImg.Bounds().Dx()
	backgroundImgHeight := g.BackgroundImg.Bounds().Dy()

	scaleX := float64(screenWidth) / float64(backgroundImgWidth)
	scaleY := float64(screenHeight) / float64(backgroundImgHeight)

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

	game := &Game{
		BackgroundImg:          backgroundImg,
		BackgroundBuildingsImg: backgroundBuildingsImg,
		spaceshipImg1:          playerImg1,
		spaceshipImg2:          playerImg2,
		Enemy1Imgs:             enemy1,
		Enemy2Imgs:             enemy2,
		clientPlayers:          map[string]*ClientPlayer{},
		clientBullets:          map[int]*ClientBullet{},
		clientEnemyFleet:       map[string]*ClientEnemy{},
	}

	game.connectToServer()
	go game.listenForServerMessages() // not sure if this should stay as a go routine
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}
