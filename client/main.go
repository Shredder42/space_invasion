package main

import (
	// "fmt"
	"image"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	scalePlayer  = 1.0 / 16.0
	scaleEnemy   = 4.0
	cooldown     = 300 * time.Millisecond
)

type Game struct {
	BackgroundImg          *ebiten.Image
	BackgroundBuildingsImg *ebiten.Image
	player                 *Player
	enemy                  *Enemy
	bullets                []Bullet
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
	shootTime time.Time
}

type Bullet struct {
	*Sprite
}

type Enemy struct {
	// *Sprite
	X            float64
	Y            float64
	speedInTps   int
	frameCounter int
	health       int64
	frame        int
	animations   map[int]*ebiten.Image
}

func (e *Enemy) Animate() {
	e.frameCounter -= 1
	if e.frameCounter < 0 {
		e.frameCounter = e.speedInTps
		if e.frame == 1 {
			e.frame = 2
		} else if e.frame == 2 {
			e.frame = 1
		}
	}
}

func (g *Game) Update() error {

	g.enemy.Animate()

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += 2
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		if g.player.shootTime.Before(time.Now().Add(-cooldown)) {
			g.player.shootTime = time.Now()
			// make a create bullet function - make sense to make a bullets package - maybe just own file?
			newBullet := Bullet{
				Sprite: &Sprite{
					Img: ebiten.NewImage(3, 6),
					// probably should make these dynamic if possible
					X: g.player.X + 16,
					Y: g.player.Y - 6,
				},
			}
			newBullet.Img.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

			g.bullets = append(g.bullets, newBullet)
			// fmt.Println(g.bullets)
		}
	}

	i := 0
	for _, bullet := range g.bullets {
		if bullet.Y > -4 {
			bullet.Y -= 4
			g.bullets[i] = bullet
			i++
		}
	}

	g.bullets = g.bullets[:i]

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

	opts.GeoM.Translate(g.player.X, g.player.Y)

	screen.DrawImage(g.player.Img, &opts)

	opts.GeoM.Reset()

	opts.GeoM.Scale(scaleEnemy, scaleEnemy)
	opts.GeoM.Translate(g.enemy.X, g.enemy.Y)
	screen.DrawImage(g.enemy.animations[g.enemy.frame], &opts)

	opts.GeoM.Reset()

	for _, bullet := range g.bullets {
		opts.GeoM.Translate(bullet.X, bullet.Y)
		screen.DrawImage(bullet.Img, &opts)
		opts.GeoM.Reset()
	}

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

	spritesImg, _, err := ebitenutil.NewImageFromFile("assets/SpaceInvaders.png")
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
		enemy: &Enemy{
			// Sprite: &Sprite{
			// 	// Img: spritesImg.SubImage(image.Rect(2, 4, 14, 12)).(*ebiten.Image),
			// 	X: 50,
			// 	Y: 50,
			// },
			// Img2:         spritesImg.SubImage(image.Rect(18, 4, 30, 16)).(*ebiten.Image),
			X:            50,
			Y:            50,
			speedInTps:   20,
			frameCounter: 20,
			health:       1,
			frame:        1,
			animations: map[int]*ebiten.Image{
				1: spritesImg.SubImage(image.Rect(2, 4, 14, 12)).(*ebiten.Image),
				2: spritesImg.SubImage(image.Rect(18, 4, 30, 14)).(*ebiten.Image),
			},
		},
	}

	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
