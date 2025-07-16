package main

import (
	// "fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	BackgroundImg          *ebiten.Image
	BackgroundBuildingsImg *ebiten.Image
	player                 *Player
}

type Sprite struct {
	Img *ebiten.Image
	X   float64
	Y   float64
	Dx  float64
	Dy  float64
}

type Player struct {
	*Sprite
}

const (
	screenWidth  = 640
	screenHeight = 480
	scalePlayer  = 1.0 / 16.0
)

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += 2
	}

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

	opts.GeoM.Scale(scalePlayer, scalePlayer)

	// playerStartX := float64(screenWidth)/2.0 - float64(g.player.Img.Bounds().Dx())*scalePlayer/2.0
	// playerStartY := float64(screenHeight) - float64(g.player.Img.Bounds().Dy())*scalePlayer
	// fmt.Println(float64(g.player.Img.Bounds().Dx()) / 2.0)
	opts.GeoM.Translate(g.player.X, g.player.Y)

	screen.DrawImage(g.player.Img, &opts)

	opts.GeoM.Reset()

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Invasion!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// fmt.Println(float64(screenWidth) / 2.0 - )

	backgroundImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_Background.png")
	if err != nil {
		log.Fatal(err)
	}

	backgroundBuildingsImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_BackgroundBuildings.png")
	if err != nil {
		log.Fatal(err)
	}

	playerImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders_white_spaceship.png")
	if err != nil {
		log.Fatal(err)
	}

	game := Game{
		BackgroundImg:          backgroundImg,
		BackgroundBuildingsImg: backgroundBuildingsImg,
		player: &Player{
			Sprite: &Sprite{
				Img: playerImg,
				X:   float64(screenWidth)/2.0 - 512.0*scalePlayer/2.0,
				Y:   float64(screenHeight) - 512.0*scalePlayer,
			},
		},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
