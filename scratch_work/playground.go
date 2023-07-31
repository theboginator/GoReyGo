package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/image/colornames"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 1024
	NumEnemies   = 3   //Number of enemies that will spawn
	NumBonus     = 3   //Number of coin collections needed to earn a bonus life
	StartHealth  = 3   //Player's starting lives
	MaxLife      = 5   //Max life the player can earn
	GoldScore    = 50  //Bonus for collecting gold
	EnemyScore   = 100 //Score for killing an enemy
	BossScore    = 200 //Score for killing the boss
)

type Sprite struct {
	pict   *ebiten.Image
	xLoc   int
	yLoc   int
	dx     int
	dy     int
	width  int
	height int
}

type Game struct {
	playerSprite   Player
	playerOrdnance Ordnance
	activeEnemies  int
	enemySprites   [NumEnemies]Enemy
	coinSprites    [NumBonus]Coins
	bossSprite     Enemy
	lifeCounter    Sprite
	enemyCounter   Sprite
	VictoryBanner  Sprite
	LossBanner     Sprite
	heart          [MaxLife]Sprite
	badguy         [NumEnemies]Sprite
	drawOps        ebiten.DrawImageOptions
	activeOrdnance bool
	collectedGold  bool
	bossDefeated   bool
}

type Player struct {
	name     string
	impact   bool
	health   int
	startX   int
	startY   int
	score    int32
	manifest Sprite
}

type Enemy struct {
	name     string
	lastMove time.Time
	health   int
	defeated bool
	startX   int
	startY   int
	manifest Sprite
	visible  bool
}

type Coins struct {
	collected bool
	lastSpawn time.Time
	manifest  Sprite
	visible   bool
}

type Ordnance struct {
	manifest Sprite
	consumed bool
}

func getPlayerInput(game *Game) { //Handle any movement from the player, and initiate any object the player fires
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		game.playerSprite.manifest.dx = -3
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		game.playerSprite.manifest.dx = 3
	} else if inpututil.IsKeyJustReleased(ebiten.KeyA) {
		game.playerSprite.manifest.dx = 0
	} else if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		game.playerSprite.manifest.dx = 0
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		game.playerSprite.manifest.dy = -3
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		game.playerSprite.manifest.dy = 3
	} else if inpututil.IsKeyJustReleased(ebiten.KeyW) {
		game.playerSprite.manifest.dy = 0
	} else if inpututil.IsKeyJustReleased(ebiten.KeyS) {
		game.playerSprite.manifest.dy = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		game.activeOrdnance = true
		launchPlayerOrdnance(game)
	}
}

func trackPlayer(game *Game) { //Move the player per keyboard input
	tmpX := game.playerSprite.manifest.xLoc
	tmpY := game.playerSprite.manifest.yLoc
	game.playerSprite.manifest.yLoc += game.playerSprite.manifest.dy
	game.playerSprite.manifest.xLoc += game.playerSprite.manifest.dx
	if game.playerSprite.manifest.xLoc != tmpX && game.playerSprite.manifest.yLoc != tmpY { //If we moved from the previous position
		game.playerSprite.impact = false //We might not be touching what we were before
	}
	for i := range game.enemySprites { //Check for a potential collision with an enemy
		if !game.enemySprites[i].defeated {
			if madeContact(game.playerSprite.manifest, game.enemySprites[i].manifest) && !game.playerSprite.impact { //If the player touches an enemy we weren't already touching
				fmt.Println("You got hit by an enemy :(")
				game.playerSprite.health -= 1
				game.playerSprite.impact = true
			}
		}
		if !game.bossDefeated && game.bossSprite.visible {
			if madeContact(game.playerSprite.manifest, game.bossSprite.manifest) && !game.playerSprite.impact {
				fmt.Println("You got hit by the boss :(")
				game.playerSprite.health -= 1
				game.playerSprite.impact = true
			}
		}
	}

}

func trackOrdnance(game *Game) { //Move any object in the direction it was fired in
	game.playerOrdnance.manifest.yLoc += game.playerOrdnance.manifest.dy
	game.playerOrdnance.manifest.xLoc += game.playerOrdnance.manifest.dx
	if game.playerOrdnance.manifest.xLoc > ScreenWidth || game.playerOrdnance.manifest.xLoc < 0 || game.playerOrdnance.manifest.yLoc > ScreenHeight || game.playerOrdnance.manifest.yLoc < 0 { //If we've hit a border
		game.activeOrdnance = false //Eliminate the object
	}
	for i := range game.coinSprites {
		if game.coinSprites[i].visible {
			if madeContact(game.playerOrdnance.manifest, game.coinSprites[i].manifest) {
				fmt.Println("You collected gold!")
				game.playerSprite.score += GoldScore
				game.coinSprites[i].visible = false
				if i+1 >= NumBonus {
					//We collected all the coins, award the extra life and restart the coin bonus process
					if game.playerSprite.health < MaxLife {
						game.playerSprite.health += 1 //Only let the player have up to 5 lives at once
					}
					fmt.Println("Another life earned! Resetting coin bonus. You now have ", game.playerSprite.health)
					loadCoins(game)
				} else {
					game.coinSprites[i+1].visible = true
				}
				game.activeOrdnance = false
			}
		}
	}
	for i := range game.enemySprites {
		if !game.enemySprites[i].defeated {
			if madeContact(game.playerOrdnance.manifest, game.enemySprites[i].manifest) { //If the object touched an enemy
				fmt.Println("You n00ned an enemy!")
				game.enemySprites[i].defeated = true
				game.activeOrdnance = false
				game.activeEnemies--
				game.playerSprite.score += EnemyScore
			}
		}
	}
	if !game.bossDefeated && game.bossSprite.visible { //hit him if we can see him
		if madeContact(game.playerOrdnance.manifest, game.bossSprite.manifest) {
			fmt.Println("You hit the boss!")
			game.bossSprite.health--
			game.activeOrdnance = false
		}
		if game.bossSprite.health == 0 {
			fmt.Println("You win! Huzzah")
			fmt.Println("Your score: ", game.playerSprite.score)
			game.bossDefeated = true
		}

	}
}

func launchPlayerOrdnance(game *Game) { //Initiate the launch of player object
	pict, _, err := ebitenutil.NewImageFromFile("assets/homework.png")
	if err != nil {
		log.Fatal("failed to load ammunition image", err)
	}
	game.playerOrdnance.manifest.pict = pict
	if game.playerSprite.manifest.dx == 0 && game.playerSprite.manifest.dy == 0 { //Launch object to the right along the x-axis if the player is stationary.
		game.playerOrdnance.manifest.dx = 6
		game.playerOrdnance.manifest.dy = 0
	} else {
		game.playerOrdnance.manifest.dx = game.playerSprite.manifest.dx * 2
		game.playerOrdnance.manifest.dy = game.playerSprite.manifest.dy * 2 //Set the direction to fire new object
	}
	game.playerOrdnance.manifest.xLoc = game.playerSprite.manifest.xLoc //Set the start point for new object to the player's current position
	game.playerOrdnance.manifest.yLoc = game.playerSprite.manifest.yLoc
	game.playerOrdnance.consumed = false
	fmt.Println("Fired object. Coords: ", game.playerOrdnance.manifest.dx, game.playerOrdnance.manifest.dy)
}

func madeContact(manifestA Sprite, manifestB Sprite) bool { //Check if 2 sprite objects are in a collision condition
	aWidth, aHeight := manifestA.pict.Size()
	bWidth, bHeight := manifestB.pict.Size()
	if manifestA.xLoc < manifestB.xLoc+aWidth &&
		manifestA.xLoc+bWidth > manifestB.xLoc &&
		manifestA.yLoc < manifestB.yLoc+aHeight &&
		manifestA.yLoc+bHeight > manifestB.yLoc {
		return true
	}
	return false
}

func (game *Game) Update() error {
	getPlayerInput(game)
	trackPlayer(game)
	if game.activeOrdnance {
		trackOrdnance(game)
	}
	if game.activeEnemies == 0 { //Enable viewing boss sprite if the other enemies are gone
		game.bossSprite.visible = true
	}

	return nil
}

func (game Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Magenta)
	game.drawOps.GeoM.Reset() //Draw the player
	game.drawOps.GeoM.Translate(float64(game.playerSprite.manifest.xLoc), float64(game.playerSprite.manifest.yLoc))
	screen.DrawImage(game.playerSprite.manifest.pict, &game.drawOps)

	game.drawOps.GeoM.Reset() //Draw the health bar
	game.drawOps.GeoM.Translate(float64(game.lifeCounter.xLoc), float64(game.lifeCounter.yLoc))
	screen.DrawImage(game.lifeCounter.pict, &game.drawOps)
	for i := 0; i < game.playerSprite.health; i++ {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.heart[i].xLoc), float64(game.heart[i].yLoc))
		screen.DrawImage(game.heart[i].pict, &game.drawOps)
	}
	game.drawOps.GeoM.Reset() //Draw the bad guys remaining bar
	game.drawOps.GeoM.Translate(float64(game.enemyCounter.xLoc), float64(game.enemyCounter.yLoc))
	screen.DrawImage(game.enemyCounter.pict, &game.drawOps)
	for i := 0; i < game.activeEnemies; i++ {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.badguy[i].xLoc), float64(game.badguy[i].yLoc))
		screen.DrawImage(game.badguy[i].pict, &game.drawOps)
	}

	if game.activeOrdnance { //Draw any object
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.playerOrdnance.manifest.xLoc), float64(game.playerOrdnance.manifest.yLoc))
		screen.DrawImage(game.playerOrdnance.manifest.pict, &game.drawOps)
	}
	for i := range game.coinSprites { //For each coin in the coin sprite array
		if game.coinSprites[i].visible { //Draw it if it's visible
			game.drawOps.GeoM.Reset()
			game.drawOps.GeoM.Translate(float64(game.coinSprites[i].manifest.xLoc), float64(game.coinSprites[i].manifest.yLoc))
			screen.DrawImage(game.coinSprites[i].manifest.pict, &game.drawOps)
		}
	}
	for i := range game.enemySprites { //For each enemy in the enemy sprite array
		if !game.enemySprites[i].defeated { //Draw the undefeated ones
			game.drawOps.GeoM.Reset()
			x := float64(game.enemySprites[i].manifest.xLoc)
			y := float64(game.enemySprites[i].manifest.yLoc)
			game.drawOps.GeoM.Translate(x, y)
			screen.DrawImage(game.enemySprites[i].manifest.pict, &game.drawOps)
		}
	}
	if game.bossSprite.visible && !game.bossDefeated {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.bossSprite.manifest.xLoc), float64(game.bossSprite.manifest.yLoc))
		screen.DrawImage(game.bossSprite.manifest.pict, &game.drawOps)
	}

	if game.bossDefeated {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.VictoryBanner.xLoc), float64(game.VictoryBanner.yLoc))
		screen.DrawImage(game.VictoryBanner.pict, &game.drawOps)
	}
	if game.playerSprite.health < 0 {
		game.drawOps.GeoM.Reset()
		game.drawOps.GeoM.Translate(float64(game.LossBanner.xLoc), float64(game.LossBanner.yLoc))
		screen.DrawImage(game.LossBanner.pict, &game.drawOps)
	}

}

func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func loadPlayer(game *Game) {
	game.playerSprite.manifest.yLoc = ScreenHeight / 2 //Setting player start point
	pict, _, err := ebitenutil.NewImageFromFile("assets/player.png")
	if err != nil {
		log.Fatal("failed to load player image", err)
	}
	game.playerSprite.manifest.pict = pict
	game.playerSprite.health = StartHealth
}

func loadEnemies(game *Game) {
	game.activeEnemies = NumEnemies
	pict, _, err := ebitenutil.NewImageFromFile("assets/enemy.png")
	if err != nil {
		log.Fatal("Failed to load enemy image", err)
	}
	rand.Seed(int64(time.Now().Second()))
	for i := range game.enemySprites {
		game.enemySprites[i].manifest.pict = pict
		game.enemySprites[i].health = 2
		game.enemySprites[i].defeated = false
		width, height := game.enemySprites[i].manifest.pict.Size()
		game.enemySprites[i].manifest.xLoc = rand.Intn(ScreenWidth - width)
		game.enemySprites[i].manifest.yLoc = rand.Intn(ScreenHeight - height)
		for madeContact(game.playerSprite.manifest, game.enemySprites[i].manifest) { //Respawn the enemy if it spawned on top of the player.
			game.enemySprites[i].manifest.xLoc = rand.Intn(ScreenWidth - width)
			game.enemySprites[i].manifest.yLoc = rand.Intn(ScreenHeight - height)
		}
	}
	pict, _, err = ebitenutil.NewImageFromFile("assets/boss.png")
	if err != nil {
		log.Fatal("Failed to load enemy image", err)
	}
	game.bossSprite.manifest.pict = pict
	game.bossSprite.health = 2
	game.bossSprite.defeated = false
	game.bossSprite.visible = false
	width, height := game.bossSprite.manifest.pict.Size()
	game.bossSprite.manifest.xLoc = rand.Intn(ScreenWidth - width)
	game.bossSprite.manifest.yLoc = rand.Intn(ScreenHeight - height)
}

func loadCoins(game *Game) {
	pict, _, err := ebitenutil.NewImageFromFile("assets/gold-coins.png")
	if err != nil {
		log.Fatal("failed to load image", err)
	}
	width, height := pict.Size()
	for i := range game.coinSprites {
		game.coinSprites[i].manifest.pict = pict
		game.coinSprites[i].collected = false
		game.coinSprites[i].visible = false
		game.coinSprites[i].manifest.xLoc = rand.Intn(ScreenWidth - width)
		game.coinSprites[i].manifest.yLoc = rand.Intn(ScreenHeight - height)
	}
	game.coinSprites[0].visible = true
}

func loadTrackers(game *Game) {
	pict, _, err := ebitenutil.NewImageFromFile("assets/lives-remain.png")
	if err != nil {
		log.Fatal("failed to load image", err)
	}
	game.lifeCounter.pict = pict
	game.lifeCounter.width, game.lifeCounter.height = pict.Size()
	game.lifeCounter.xLoc = 0
	game.lifeCounter.yLoc = ScreenHeight - game.lifeCounter.height
	pict, _, err = ebitenutil.NewImageFromFile("assets/heart.png")
	if err != nil {
		log.Fatal("failed to load image, ", err)
	}
	for i := range game.heart {
		game.heart[i].pict = pict
		game.heart[i].width, game.heart[i].height = pict.Size()
		game.heart[i].xLoc = (game.heart[i].width * i) + game.lifeCounter.width
		game.heart[i].yLoc = ScreenHeight - game.heart[i].height
	}
	pict, _, err = ebitenutil.NewImageFromFile("assets/enemies-remain.png")
	if err != nil {
		log.Fatal("Failed to load image", err)
	}
	game.enemyCounter.pict = pict
	game.enemyCounter.width, game.enemyCounter.height = game.enemyCounter.pict.Size()
	game.enemyCounter.xLoc = 0
	game.enemyCounter.yLoc = (ScreenHeight - game.enemyCounter.height) - game.lifeCounter.height //Stack the enemy life counter on top of the player life counter
	pict, _, err = ebitenutil.NewImageFromFile("assets/enemy-count.png")
	if err != nil {
		log.Fatal("failed to load image, ", err)
	}
	for i := range game.badguy {
		game.badguy[i].pict = pict
		game.badguy[i].width, game.badguy[i].height = pict.Size()
		game.badguy[i].xLoc = (game.badguy[i].width * i) + game.enemyCounter.width
		game.badguy[i].yLoc = (ScreenHeight - game.heart[i].height) - game.lifeCounter.height
	}
}

func loadBanners(game *Game) {
	game.VictoryBanner.yLoc = ScreenHeight / 2 //Setting player start point
	pict, _, err := ebitenutil.NewImageFromFile("assets/winning-screen.png")
	if err != nil {
		log.Fatal("failed to load image", err)
	}
	game.VictoryBanner.pict = pict
	game.LossBanner.yLoc = ScreenHeight / 2 //Setting player start point
	pict, _, err = ebitenutil.NewImageFromFile("assets/nooned.png")
	if err != nil {
		log.Fatal("failed to load image", err)
	}
	game.LossBanner.pict = pict
}

func setup_database() *sql.DB { //Create the database
	database, _ := sql.Open("sqlite3", "./DeathRoadScores.db")
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS highscores (id INTEGER PRIMARY KEY, username TEXT, score INTEGER)")
	statement.Exec()
	return database
}

func main() {
	database := setup_database()
	keyboard := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a really cool username: ")
	uname, err := keyboard.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("The Death Road Through DMF")
	gameObject := Game{}
	loadTrackers(&gameObject)
	loadPlayer(&gameObject)
	loadEnemies(&gameObject)
	loadCoins(&gameObject)
	loadBanners(&gameObject)
	if err := ebiten.RunGame(&gameObject); err != nil {
		log.Fatal("Oh no! something terrible happened", err)
	}
	fmt.Println("Inserting ", uname, " with score ", gameObject.playerSprite.score)
	statement, err := database.Prepare("INSERT INTO highscores (username, score) VALUES (?, ?)")

	//TODO: Sanitize inputs before insertion
	statement.Exec(string(uname), int32(gameObject.playerSprite.score))
	if err != nil {
		log.Fatal("Something terrible happened. ", err)
	}
}
