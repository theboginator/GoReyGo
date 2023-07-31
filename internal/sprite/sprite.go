package sprite

import (
	"GoReyGo/internal/world"
	"github.com/hajimehoshi/ebiten/v2"
)

type Vector struct {
	dx int
	dy int
}

type Sprite struct {
	Pict            *ebiten.Image
	CurrentPos      world.Location
	SpawnPos        world.Location
	CurrentVelocity Vector
	Width           int
	Height          int
	Health          int
	Score           int
	Name            string
}
