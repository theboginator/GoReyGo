package main

import (
	"GoReyGo/internal/game"
	"GoReyGo/internal/pal"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	pal.PrintLog(pal.Info, "Starting GoReyGo")
	pal.PrintLog(pal.Warn, "Test warning message")
	pal.PrintLog(pal.Error, "Test error message")

	/*
		keyboard := bufio.NewReader(os.Stdin)
		fmt.Println("Enter a really cool username: ")
		userName, err := keyboard.ReadString('\n')
		if err != nil {
			pal.PrintLog(pal.Error, fmt.Sprintf("%s", err))
		}
		userName = strings.Trim(userName, "\n")

		pal.PrintLog(pal.Info, fmt.Sprintf("User %s joined the game", userName))
	*/
	ebiten.SetWindowSize(pal.ScreenWidth, pal.ScreenHeight)
	ebiten.SetWindowTitle("Rey The Fox!")

	var g game.Game
	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}
