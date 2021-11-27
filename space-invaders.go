package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Defining the structs

type Game struct {}

type Cell struct {
	x int
	y int
}

type Alien struct {
  id int
  pos Cell
  sprite *ebiten.Image
}

type Player struct {
  x float64
  y float64
}

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
  player = Player{x: WindowW/2-10, y: WindowH-25}  
  alienSleepMS = 1000
  usableGridCells int
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

  totalCells := rows * cols

  totalUnusableRows := TopUnusableRows + BotUnusableRows
  unusableCellsOfRows := totalUnusableRows * cols
  unusableSideCells := (rows - totalUnusableRows) * SideUnusableCols * 2

  usableGridCells = totalCells - unusableCellsOfRows - unusableSideCells
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

  grid[alien.pos.y][alien.pos.x] = 0
  grid[target.y][target.x] = alien.id
  alien.pos = target


}

func initAliens(amount int){
  // Alien maximum value is 50% of the grid, this logic may be moved later
  cappedAmount := int(math.Min(float64(amount),float64(usableGridCells/2)))

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

  fmt.Printf("Generated %d aliens\n",len(aliens))

  go alienBrain(1)
}

func drawAliens(screen *ebiten.Image){
  for i:= range aliens{
    options := new(ebiten.DrawImageOptions)
    options.GeoM.Translate(float64(aliens[i].pos.x*CellSize),float64(aliens[i].pos.y*CellSize))
    screen.DrawImage(aliens[i].sprite,options)
  }
}

func (g *Game) Update() error {
  if ebiten.IsKeyPressed(ebiten.KeyRight) {
    player.x += 5
  } 
  if ebiten.IsKeyPressed(ebiten.KeyLeft) {
    player.x -= 5
  }
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  opts := &ebiten.DrawImageOptions{}
  opts.GeoM.Translate(player.x, player.y)
  img := ebiten.NewImage(20, 20)
  img.Fill(color.RGBA{0, 255, 0, 255})
  screen.DrawImage(img, opts)
  drawAliens(screen)
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

  initAliens(1)


  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}