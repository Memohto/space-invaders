package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
  TopUnusableRows = 2
  BotUnusableRows = 3
  SideUnusableCols = 1 // Unusable cols on the left and right
)

var (
  grid [][]int
  aliens []Alien
  alienSprite *ebiten.Image
  alienSleepMS = 1000
  accessGrid chan bool
  oX,oY,oXf,oYf float64
)

// Helper functions

func isSameCell(cell1, cell2 Cell) bool{
  return cell1.x == cell2.x && cell1.y == cell2.y
}

func cellIsInBounds(cell Cell) bool {
  return cell.x>=0 && cell.y>=0 && cell.x<len(grid[0]) && cell.y<len(grid)
}

func cellIsFree(cell Cell) bool{
  return cellIsInBounds(cell) && grid[cell.y][cell.x] == 0
}

func disableGridRow(row int){
  for i:= range grid[row]{
    grid[row][i]=-1
  }
}

// Game logic

func initGrid(){
  rows := int(math.Floor(WindowH/CellSize))
  cols := int(math.Floor(WindowW/CellSize))
  
  fmt.Printf("%dx%d grid created\n",cols,rows)
  
  grid = make([][]int, rows)

  topUnusable := TopUnusableRows
  for i := range grid {
    grid[i] = make([]int, cols)
    if topUnusable>0 {
      topUnusable--
      disableGridRow(i)
    }else if rows-1-i < BotUnusableRows{
      disableGridRow(i)
    }else{
      for j := range grid[i] {
        if j<SideUnusableCols || cols-1-j<SideUnusableCols{
          grid[i][j]=-1
        }
      }
    }
  }

  offset := CellSize/2.0
  oX = CellSize * SideUnusableCols - offset
  oXf = WindowW - CellSize*SideUnusableCols + offset
  oY = CellSize * TopUnusableRows - offset
  oYf = WindowH-CellSize*BotUnusableRows + offset
}

func alienBrain(id int){
  myAlien := &aliens[id-1]

  for {
    time.Sleep(time.Duration(alienSleepMS) * time.Millisecond)
    moveAlien(myAlien)
  }
}

func moveAlien(alien *Alien){// Moves alien randomly (Down, Left or Right)
  var target Cell

  for noMovement := true; noMovement; {// Do while
    target = alien.pos
    switch rand.Intn(4) {
      case 0:// Down
        target.y++
      case 1:// Left
        target.x--
      case 2:// Right
        target.x++
      case 3:// Don't move
        noMovement=false
    }
    
    if isSameCell(target,alien.pos) || cellIsFree(target) {
      noMovement = false
    }

  }

  <- accessGrid

  if cellIsFree(target) {
    //If another Alien took the target position in the time frame
    // between calculation and execution, we surrender the space
    grid[alien.pos.y][alien.pos.x] = 0
    grid[target.y][target.x] = alien.id
    alien.pos = target
  }

  accessGrid <- true


}

func initAliens(amount int){
  // Alien maximum value is 50% of the grid, this logic may be moved later
  maxCells := ((len(grid) - (BotUnusableRows + TopUnusableRows))/2) * (len(grid[0])-SideUnusableCols*2)
  
  cappedAmount := int(math.Min(float64(amount),float64(maxCells)))

  // Sprite
  alienSprite = ebiten.NewImage(CellSize,CellSize)
  debugColor:= color.RGBA{uint8(255), 0, 0, uint8(255)}
  alienSprite.Fill(debugColor)

  aliens = make([]Alien, cappedAmount)

  curRow := TopUnusableRows
  curCol := SideUnusableCols

  for i:= range aliens{
    aliens[i] = Alien{
      id:i+1,
      pos: Cell{x:curCol,y:curRow},
      sprite: alienSprite,
    }
    grid[curRow][curCol] = i+1
    if curCol>0 && curCol%(len(grid[0])-1-SideUnusableCols) == 0 {
      curRow++
      curCol = SideUnusableCols
    }else{
      curCol++
    }
  }

  for i:= range aliens {
    go alienBrain(aliens[i].id)
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

func drawOverlay(screen *ebiten.Image){
  ebitenutil.DrawLine(screen,oX,oY,oXf,oY,color.White)
  ebitenutil.DrawLine(screen,oX,oY,oX,oYf,color.White)
  ebitenutil.DrawLine(screen,oXf,oY,oXf,oYf,color.White)
  ebitenutil.DrawLine(screen,oX,oYf,oXf,oYf,color.White)
}

func (g *Game) Update() error {
  // Write your game's logical update.
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  drawAliens(screen)
  drawOverlay(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return WindowW, WindowH
}

func main() {
  game := &Game{}
  ebiten.SetWindowSize(WindowW, WindowH)
  ebiten.SetWindowTitle("Space Invaders")
  
  rand.Seed(time.Now().UnixNano())

  initGrid()

  accessGrid = make( chan bool , 1)
  accessGrid <- true

  initAliens(20)


  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}