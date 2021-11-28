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

type Game struct {}

type Cell struct {
	x int
	y int
}

type Alien struct {
  id int
  dead bool
  pos Cell
  sprite *ebiten.Image
}

type Player struct {
  lives int
  pos Cell
  sprite *ebiten.Image
}

type Bullet struct {
  ready bool
  pos Cell
  sprite *ebiten.Image
}

// Global variables

const (
  WindowH = 640
  WindowW = 420
  CellSize = 20
  TopUnusableRows = 2
  BotUnusableRows = 3
  SideUnusableCols = 1 // Unusable cols on the left and right
  MaxBullets = 100
  BulletXOffset = 8
  BulletYOffset = 3
)

var (
  gameState int // 0 = on going, 1 = win, 2 = lose
  score int
  grid [][]int
  aliens []Alien
  alienSprite *ebiten.Image
  alienSleepMS = 1000
  accessGrid chan bool
  oX,oY,oXf,oYf float64
  player Player
  playerSprite *ebiten.Image
  playerMoveSleepMS = 150
  playerShootSleepMS = 1000
  bullets []Bullet
  bulletSprite  *ebiten.Image
  bulletSleepMS = 50
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

func checkCollision(gridValue int) {
  if gridValue > 0 {
    for i, a := range aliens {
      if a.id == gridValue {
        aliens[i].dead = true
        grid[a.pos.y][a.pos.x] = 0
        score += 1
        break
      }
    }
  }
}

func countAliens() int {
  count := 0
  for _, a := range aliens {
    if !a.dead {
      count += 1
    }
  }
  return count
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
  myAlien := &aliens[id-2]

  for !myAlien.dead && gameState == 0{
    time.Sleep(time.Duration(alienSleepMS) * time.Millisecond)
    moveAlien(myAlien)
  }
}

func moveAlien(alien *Alien){// Moves alien randomly (Down, Left or Right)
  var target Cell

  for noMovement := true; noMovement && !alien.dead; {// Do while
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
  maxCells := ((len(grid) - (BotUnusableRows + TopUnusableRows))/4) * (len(grid[0])-SideUnusableCols*2)
  
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
      id:i+2,
      pos: Cell{x:curCol,y:curRow},
      sprite: alienSprite,
    }
    grid[curRow][curCol] = aliens[i].id
    if curCol>0 && curCol%(len(grid[0])-1-SideUnusableCols) == 0 {
      curRow+=2
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
  for i, a := range aliens {
    if !a.dead {
      options := new(ebiten.DrawImageOptions)
      options.GeoM.Translate(float64(aliens[i].pos.x*CellSize),float64(aliens[i].pos.y*CellSize))
      screen.DrawImage(aliens[i].sprite,options)
    }
  }
}

func moveBullet(bullet *Bullet, direction int) {
  for bullet.ready && cellIsFree(Cell{bullet.pos.x, bullet.pos.y-1}) {
    grid[bullet.pos.y][bullet.pos.x] = 0
    bullet.pos.y += direction
    grid[bullet.pos.y][bullet.pos.x] = -2
    time.Sleep(time.Duration(bulletSleepMS) * time.Millisecond)
  }
  bullet.ready = false
  grid[bullet.pos.y][bullet.pos.x] = 0
  checkCollision(grid[bullet.pos.y-1][bullet.pos.x])
}

func initBullets() {
  bullets = make([]Bullet, MaxBullets)

  xSize := CellSize - BulletXOffset*2
  ySize := CellSize - BulletYOffset*2
  bulletSprite = ebiten.NewImage(xSize,ySize)
  debugColor := color.RGBA{uint8(255), uint8(255), uint8(255), uint8(255)}
  bulletSprite.Fill(debugColor)
}

func drawBullets(screen *ebiten.Image) {
  for _, b := range bullets {
    if b.ready {
      options := new(ebiten.DrawImageOptions)
      options.GeoM.Translate(float64(b.pos.x*CellSize+BulletXOffset), float64(b.pos.y*CellSize+BulletYOffset))
      screen.DrawImage(b.sprite, options)
    }
  }
}

func playerMove() {
  for player.lives > 0 && gameState == 0 {
    time.Sleep(time.Duration(playerMoveSleepMS) * time.Millisecond)
    if ebiten.IsKeyPressed(ebiten.KeyRight) && cellIsFree(Cell{x:player.pos.x + 1,y:player.pos.y}) {
      grid[player.pos.y][player.pos.x] = 0; 
      player.pos.x += 1
      grid[player.pos.y][player.pos.x] = 1; 
    } 
    if ebiten.IsKeyPressed(ebiten.KeyLeft) && cellIsFree(Cell{x:player.pos.x - 1,y:player.pos.y}){
      grid[player.pos.y][player.pos.x] = 0; 
      player.pos.x -= 1
      grid[player.pos.y][player.pos.x] = 1; 
    }
  }
}

func playerShoot() {
  for player.lives > 0 && gameState == 0 {
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
      generatePlayerBullet()
      time.Sleep(time.Duration(playerShootSleepMS) * time.Millisecond)
    } 
  }
}

func generatePlayerBullet() {
  for i, b := range bullets {
    if !b.ready {
      xPos := player.pos.x
      yPos := player.pos.y-1
      grid[yPos][xPos] = -2
      bullets[i] = Bullet{
        ready: true,
        pos: Cell{x:xPos, y:yPos},
        sprite: bulletSprite,
      }
      go moveBullet(&bullets[i], -1)
      break
    }
  }
}

func initPlayer() {
  playerSprite = ebiten.NewImage(CellSize,CellSize)
  debugColor:= color.RGBA{0, uint8(255), 0, uint8(255)}
  playerSprite.Fill(debugColor)
  
  xPos := (WindowW/CellSize - SideUnusableCols*2) / 2
  yPos := WindowH/CellSize - BotUnusableRows - 1
  player = Player{
    lives: 10, 
    pos: Cell{x:xPos,y:yPos},
    sprite: playerSprite,
  }

  go playerMove()
  go playerShoot()
}

func drawPlayer(screen *ebiten.Image) {
  options := new(ebiten.DrawImageOptions)
  options.GeoM.Translate(float64(player.pos.x*CellSize), float64(player.pos.y*CellSize))
  screen.DrawImage(player.sprite, options)
}

func drawText(screen *ebiten.Image) {
  yPos := WindowH-CellSize*2
  scoreString := fmt.Sprintf("%d", score)
  livesString := fmt.Sprintf("%d", player.lives)
  ebitenutil.DebugPrintAt(screen, "Score: "+scoreString, CellSize, yPos)
  ebitenutil.DebugPrintAt(screen, "Lives: "+livesString, WindowW-CellSize*4, yPos)
  if gameState == 1 {
    ebitenutil.DebugPrintAt(screen, "You win", WindowW/2-CellSize*1, WindowH/2)
  }
  if gameState == 2 {
    ebitenutil.DebugPrintAt(screen, "Game over", WindowW/2-CellSize*2, WindowH/2)
  }
}

func drawOverlay(screen *ebiten.Image){
  ebitenutil.DrawLine(screen,oX,oY,oXf,oY,color.White)
  ebitenutil.DrawLine(screen,oX,oY,oX,oYf,color.White)
  ebitenutil.DrawLine(screen,oXf,oY,oXf,oYf,color.White)
  ebitenutil.DrawLine(screen,oX,oYf,oXf,oYf,color.White)
}

func (g *Game) Update() error {
  if countAliens() == 0 && gameState == 0 {
    gameState = 1
  } else if player.lives < 0 && gameState == 0 {
    gameState = 2
  }
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  drawBullets(screen)
  drawPlayer(screen)
  drawAliens(screen)
  drawOverlay(screen)
  drawText(screen)
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

  initBullets()
  initPlayer()
  initAliens(200)

  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}