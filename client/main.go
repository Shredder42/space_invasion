package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Shredder42/space_invasion/shared"
	"github.com/gorilla/websocket"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 900
	screenHeight = 600
	scalePlayer  = 1.0 / 16.0
	scaleEnemy   = 4.0
	cooldown     = 300 * time.Millisecond

	root  = "http://localhost:8080/"
	api   = "api/"
	users = "users"
)

type Game struct {
	BackgroundImg          *ebiten.Image
	BackgroundBuildingsImg *ebiten.Image
	spaceshipImg1          *ebiten.Image
	spaceshipImg2          *ebiten.Image
	Enemy1Imgs             map[int]*ebiten.Image
	Enemy2Imgs             map[int]*ebiten.Image
	gameFont               font.Face

	conn       *websocket.Conn
	connected  bool
	myPlayerID string

	// gameState *shared.GameState

	clientPlayers    map[string]*ClientPlayer
	clientBullets    map[int]*ClientBullet
	clientEnemyFleet map[string]*ClientEnemy
}

func getCredentials() (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("(C)reate account or (L)ogin")
	option, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading option: %v", err)
	}

	fmt.Println("Enter username: ")
	userName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading user name: %v", err)
	}

	fmt.Println("Enter password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("error reading password: %v", err)
	}

	option = strings.Trim(option, "\n")
	userName = strings.Trim(option, "\n")
	password = strings.Trim(option, "\n")

	return strings.ToLower(option), userName, password

}

type Credentials struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func createAccount(userName, password, path string) {
	credentials := Credentials{
		UserName: userName,
		Password: password,
	}

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		log.Printf("error marshaling credentials: %v", err)
		return
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(credentialsJSON))
	if err != nil {
		log.Println("error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("error making request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		log.Println("account not created")
		return
	}

	log.Printf("Account created for user: %s", userName)

}

func loadFontFace(path string, size float64) font.Face {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to load font: %v", err)
	}

	tt, err := opentype.Parse(fontBytes)
	if err != nil {
		log.Printf("text file not OpenType or TrueType: %v", err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Printf("couldn't create font.Face: %v", err)
	}

	return face
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

	// textOpts := &text.DrawOptions{}
	// textOpts.GeoM.Translate(100, 100)
	// textOpts.ColorScale.ScaleWithColor(color.White)

	for _, player := range g.clientPlayers {
		if player.ID == "player_1" {
			text.Draw(screen, fmt.Sprintf("Player 1: %d", player.Score), g.gameFont, 30, 20, color.White)
		}
		if player.ID == "player_2" {
			text.Draw(screen, fmt.Sprintf("Player 2: %d", player.Score), g.gameFont, 750, 20, color.White)
		}
	}
	// if g.myPlayerID == "player_2" {
	// 	text.Draw(screen, "player 2", g.gameFont, 500, 200, color.White)
	// }

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
	option, userName, password := getCredentials()

	if option != "c" && option != "l" {
		log.Fatal("invalid option")
	}

	if option == "c" {
		createAccount(userName, password, root+api+users)
	}

	// login

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

	gameFont := loadFontFace("assets/fonts/Roboto-Black.ttf", 24)

	game := &Game{
		BackgroundImg:          backgroundImg,
		BackgroundBuildingsImg: backgroundBuildingsImg,
		spaceshipImg1:          playerImg1,
		spaceshipImg2:          playerImg2,
		Enemy1Imgs:             enemy1,
		Enemy2Imgs:             enemy2,
		gameFont:               gameFont,
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
