package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Defining the structs
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

// Global variables
const (
  WindowH = 640
  WindowW = 420
  CellSize = 20
)

var (
  grid [][]int
  aliens []Alien
  alienSprite *ebiten.Image
)

func initGrid(){
  rows := int(math.Floor(WindowH/CellSize))
  cols := int(math.Floor(WindowW/CellSize))
  
  fmt.Printf("%dx%d grid created\n",cols,rows)
  
  grid = make([][]int, rows)
  for i := range grid {
    grid[i] = make([]int, cols)
  }
}

func initAliens(amount int){
  // Alien maximum value is 50% of the grid, this logic may be moved later
  cappedAmount := int(math.Min(float64(amount),math.Floor(float64(len(grid)*len(grid[0]))/2)))

  // Sprite
  alienSprite = ebiten.NewImage(CellSize,CellSize)
  debugColor:= color.RGBA{uint8(255), 0, 0, uint8(255)}
  alienSprite.Fill(debugColor)

  aliens = make([]Alien, cappedAmount)

  curRow := 0
  curCol := 0

  for i:= range aliens{
    aliens[i] = Alien{
      id:i,
      pos: Cell{x:curCol,y:curRow},
      sprite: alienSprite,
    }
    grid[curRow][curCol] = i
    if curCol>0 && curCol%(len(grid[0])-1) == 0 {
      curRow++
      curCol=0
    }else{
      curCol++
    }
  }

  fmt.Printf("Generated %d aliens\n",len(aliens))
}

func drawAliens(screen *ebiten.Image){
  for i:= range aliens{
    options := new(ebiten.DrawImageOptions)
    options.GeoM.Translate(float64(aliens[i].pos.x*CellSize),float64(aliens[i].pos.y*CellSize))
    screen.DrawImage(aliens[i].sprite,options)
  }
}

func (g *Game) Update() error {
  // Write your game's logical update.
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  drawAliens(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return WindowW, WindowH
}

func main() {
  game := &Game{}
  ebiten.SetWindowSize(WindowW, WindowH)
  ebiten.SetWindowTitle("Space Invaders")
  
  initGrid()
  
  initAliens(50)


  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}