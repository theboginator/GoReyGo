package game

import (
	"GoReyGo/internal/pal"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/colornames"
)

type Game struct {
	drawOps ebiten.DrawImageOptions
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return pal.ScreenWidth, pal.ScreenHeight
}

func (g *Game) Update() error {
	// This is the logic that should run on every state update
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "This is a test.")
	screen.Fill(colornames.Magenta)
	g.drawOps.GeoM.Reset()
	// screen.DrawImage()
}
