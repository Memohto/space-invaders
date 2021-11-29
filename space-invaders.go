package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"
  "strconv"
  "unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/nfnt/resize"
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
  deadTicks int
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

  AlienSleepMS = 1000
  PlayerMoveSleepMS = 150
  PlayerShootSleepMS = 1000
  BulletSleepMS = 50
  AlienBulletSleepMS = 100
  AliensDeadTicks = 15
  AlienShootProbability = 0.05
)

var (
  gameState int = 3 // 0 = on going, 1 = win, 2 = lose, 3 = menu
  score int
  grid [][]int
  aliens []Alien
  alienCount string
  accessGrid chan bool
  oX,oY,oXf,oYf float64
  player Player
  bullets []Bullet

  //Imgs
  imgLogo *ebiten.Image
  imgIcon *ebiten.Image
  imgAliens []*ebiten.Image
  imgPlayer *ebiten.Image
  imgBullet *ebiten.Image
  imgExpl *ebiten.Image
  imgAlienBullet *ebiten.Image
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

func getSubImgFrom(sheet *ebiten.Image,x,y int)*ebiten.Image{
  return sheet.SubImage(image.Rect(CellSize*x,y*CellSize,CellSize*(x+1),CellSize*(y+1))).(*ebiten.Image)
}


// Game logic

func initGrid(){
  rows := int(math.Floor(WindowH/CellSize))
  cols := int(math.Floor(WindowW/CellSize))
  
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
  oY = WindowH-CellSize*(BotUnusableRows-1) - offset
  oYf = WindowH - offset
}

func loadImages() error {
  var err error

  imgLogo,_,err = ebitenutil.NewImageFromFile("img/Logo.png")
  if err != nil {
    return err
  }

  imgIcon,_,err = ebitenutil.NewImageFromFile("img/GameIcon.png")
  if err != nil {
    return err
  }

  icon := make([]image.Image, 1)
  icon[0]=imgIcon
  ebiten.SetWindowIcon(icon)

  _,spriteSheet,err := ebitenutil.NewImageFromFile("img/SpaceInvaders.png")
  if err != nil {
    return err
  }

  resizedSpriteSheet := ebiten.NewImageFromImage(resize.Resize(CellSize*7,CellSize*5,spriteSheet,resize.NearestNeighbor))

 
  imgPlayer = getSubImgFrom(resizedSpriteSheet,4,0)
  
  imgBullet = getSubImgFrom(resizedSpriteSheet,2,0)

  imgAliens = make([]*ebiten.Image,3)
  imgAliens[0] = getSubImgFrom(resizedSpriteSheet,0,0)
  imgAliens[1] = getSubImgFrom(resizedSpriteSheet,0,1)
  imgAliens[2] = getSubImgFrom(resizedSpriteSheet,0,2)

  imgExpl = getSubImgFrom(resizedSpriteSheet,2,3)

  imgAlienBullet = getSubImgFrom(resizedSpriteSheet,2,1)

  return nil
}

func alienBrain(id int){
  myAlien := &aliens[id-2]

  for !myAlien.dead && gameState == 0 {
    time.Sleep(time.Duration(AlienSleepMS) * time.Millisecond)
    moveAlien(myAlien)
    alienShoot(myAlien)
  }
}

func alienDamage(){
  player.lives--
}

func generateAlienBullet(alien *Alien) {
  for i, b := range bullets {
    if !b.ready {
      xPos := alien.pos.x
      yPos := alien.pos.y-1
      grid[yPos][xPos] = -2
      bullets[i] = Bullet{
        ready: true,
        pos: Cell{x:xPos, y:yPos},
        sprite: imgAlienBullet,
      }
      go moveBullet(&bullets[i], 1)
      break
    }
  }
}

func alienShoot(alien *Alien){
  if rand.Float64()<AlienShootProbability{
    generateAlienBullet(alien)
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

  if cellIsFree(target) && !alien.dead {
    //If another Alien took the target position in the time frame
    // between calculation and execution, we surrender the space
    grid[alien.pos.y][alien.pos.x] = 0
    grid[target.y][target.x] = alien.id
    alien.pos = target
  }

  accessGrid <- true

  if alien.pos.y >= player.pos.y {
    alien.dead = true
    alienDamage()
  }


}

func initAliens(amount int){
  // Alien maximum value is 50% of the grid, this logic may be moved later
  maxCells := ((len(grid) - (BotUnusableRows + TopUnusableRows))/4) * (len(grid[0])-SideUnusableCols*2)
  
  cappedAmount := int(math.Min(float64(amount),float64(maxCells)))

  aliens = make([]Alien, cappedAmount)

  curRow := TopUnusableRows
  curCol := SideUnusableCols

  alienVariant := 0

  for i:= range aliens{
    aliens[i] = Alien{
      id:i+2,
      pos: Cell{x:curCol,y:curRow},
      sprite: imgAliens[alienVariant],
      deadTicks: AliensDeadTicks,
    }
    alienVariant = (alienVariant+1)%(len(imgAliens))
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
}

func drawAliens(screen *ebiten.Image){
  for i, a := range aliens {
    if !a.dead {
      options := new(ebiten.DrawImageOptions)
      options.GeoM.Translate(float64(aliens[i].pos.x*CellSize),float64(aliens[i].pos.y*CellSize))
      screen.DrawImage(aliens[i].sprite,options)
    }else if a.deadTicks>0{
      options := new(ebiten.DrawImageOptions)
      options.GeoM.Translate(float64(aliens[i].pos.x*CellSize),float64(aliens[i].pos.y*CellSize))
      screen.DrawImage(imgExpl,options)
    }
  }
}

func moveBullet(bullet *Bullet, direction int) {

  isPlayer := direction<0

  if isPlayer {
    for cellIsFree(Cell{bullet.pos.x, bullet.pos.y-1}) {
        grid[bullet.pos.y][bullet.pos.x] = 0
        bullet.pos.y += direction
        grid[bullet.pos.y][bullet.pos.x] = -2
        time.Sleep(time.Duration(BulletSleepMS) * time.Millisecond)
    }
  }else{
    for !isSameCell(player.pos,bullet.pos) && bullet.pos.y < player.pos.y {
      bullet.pos.y += direction
      time.Sleep(time.Duration(AlienBulletSleepMS) * time.Millisecond)
    }
  }

  bullet.ready = false

  if isPlayer{
    grid[bullet.pos.y][bullet.pos.x] = 0
    checkCollision(grid[bullet.pos.y-1][bullet.pos.x])
  }else if isSameCell(player.pos,bullet.pos){
    alienDamage()
  }
}

func initBullets() {
  bullets = make([]Bullet, MaxBullets)
}

func drawBullets(screen *ebiten.Image) {
  for _, b := range bullets {
    if b.ready {
      options := new(ebiten.DrawImageOptions)
      options.GeoM.Translate(float64(b.pos.x*CellSize), float64(b.pos.y*CellSize))
      screen.DrawImage(b.sprite, options)
    }
  }
}

func playerMove() {
  for player.lives > 0 && gameState == 0 {
    time.Sleep(time.Duration(PlayerMoveSleepMS) * time.Millisecond)
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
      time.Sleep(time.Duration(PlayerShootSleepMS) * time.Millisecond)
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
        sprite: imgBullet,
      }
      go moveBullet(&bullets[i], -1)
      break
    }
  }
}

func initPlayer() {
  //playerSprite = ebiten.NewImage(CellSize,CellSize)
  //debugColor:= color.RGBA{0, uint8(255), 0, uint8(255)}
  //playerSprite.Fill(debugColor)
  
  xPos := (WindowW/CellSize - SideUnusableCols*2) / 2
  yPos := WindowH/CellSize - BotUnusableRows - 1
  player = Player{
    lives: 10, 
    pos: Cell{x:xPos,y:yPos},
    sprite: imgPlayer,
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

func drawMenu(screen *ebiten.Image) {
  options := new(ebiten.DrawImageOptions)
  options.GeoM.Translate(WindowW/2-CellSize*4, WindowH/2-CellSize*10)
  options.GeoM.Scale(0.7, 0.7)
  screen.DrawImage(imgLogo,options)
  ebitenutil.DebugPrintAt(screen, "Type alien amount (max = 50, default = 20):", CellSize*4, WindowH/2)
  ebitenutil.DebugPrintAt(screen, alienCount, WindowW/2-CellSize*1, WindowH/2+CellSize)
  ebitenutil.DebugPrintAt(screen, "Press enter to start", CellSize*7, WindowH/2+CellSize*4)
}

func inputNumber() {
  var input []rune
  input = ebiten.AppendInputChars(input)
  for _, r := range input {
    if unicode.IsDigit(r) {
      alienCount += string(r)
    } else if r == 100 {
      if len(alienCount) > 0 {
        alienCount = alienCount[:len(alienCount)-1]
      }
    }
  }
}

func (g *Game) Update() error {
  if ebiten.IsKeyPressed(ebiten.KeyEnter) && (gameState > 0) {
    intCount, _ := strconv.Atoi(alienCount)
    if intCount < 1 {
      intCount = 1
    } else if intCount > 50 {
      intCount = 50
    }
    gameState = 0
    score = 0
    initGrid()
    initBullets()
    initPlayer()
    initAliens(intCount)
  }
  if countAliens() == 0 && gameState == 0 {
    gameState = 1
  } else if player.lives <= 0 && gameState == 0 {
    gameState = 2
  }
  if gameState == 3 {
    inputNumber()
  }
  for i:= range aliens{
    if aliens[i].dead && aliens[i].deadTicks>0 {
      aliens[i].deadTicks--
    }
  }
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
  if gameState == 0 {
    drawPlayer(screen)
    drawAliens(screen)
    drawBullets(screen)
  } else if gameState == 3 {
    drawMenu(screen)
  }
  drawOverlay(screen)
  drawText(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
  return WindowW, WindowH
}

func main() {
  game := &Game{}
 
	if  err := loadImages(); err != nil {
		log.Fatal(err)
	}
  ebiten.SetWindowSize(WindowW, WindowH)
  ebiten.SetWindowTitle("Space Invaders")
  
  rand.Seed(time.Now().UnixNano())

  initGrid()

  accessGrid = make( chan bool , 1)
  accessGrid <- true

  if err := ebiten.RunGame(game); err != nil {
    log.Fatal(err)
  }
}