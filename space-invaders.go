package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

//Defining the structs
type Cell struct {
	x int
	y int
}

type Alien struct {
  id int
  pos Cell
  sprite *ebiten.Image
}

type Game struct{}

//Global variables
const (
  WindowH = 640
  WindowW = 420
  CellSize = 20
)

var (
  grid [][]int //Size gets initialized on main
  alienTest Alien//This will become an array of aliens
  alienSprite *ebiten.Image
)

func initGrid(){
  rows := int(math.Floor(WindowH/CellSize))
  cols := int(math.Floor(WindowW/CellSize))
  
  fmt.Printf("%dx%d grid created\n",rows,cols)
  
  grid = make([][]int, rows)
  for i := range grid {
    grid[i] = make([]int, cols)
  }
}

func initAliens(){
  alienSprite = ebiten.NewImage(CellSize,CellSize)
  debugColor:= color.RGBA{uint8(255), 0, 0, uint8(255)}
  alienSprite.Fill(debugColor)
  
  alienTest = Alien{
    id: 0,
    pos: Cell{x:10,y:20},
    sprite: alienSprite,
  }
}

func (g *Game) Update() error {
  // Write your game's logical update.
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  options := new(ebiten.DrawImageOptions)
  options.GeoM.Translate(float64(alienTest.pos.x),float64(alienTest.pos.y))
  //ebitenutil.DrawRect(screen,float64(alienTest.pos.x),float64(alienTest.pos.y),CellSize,CellSize,color.White)
  screen.DrawImage(alienTest.sprite,options)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return 320, 240
}

func main() {
  game := &Game{}
  ebiten.SetWindowSize(WindowW, WindowH)
  ebiten.SetWindowTitle("Space Invaders")
  
  initGrid()
  
  initAliens()
  
  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}