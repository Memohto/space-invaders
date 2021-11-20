package main

import (
  "log"
	"github.com/hajimehoshi/ebiten/v2"
	// "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

func (g *Game) Update() error {
  // Write your game's logical update.
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  // Write your game's rendering.
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return 320, 240
}

func main() {
  game := &Game{}
  ebiten.SetWindowSize(640, 480)
  ebiten.SetWindowTitle("Space Invaders")
  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}