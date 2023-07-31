package object

import (
	"GoReyGo/internal/sprite"
	"GoReyGo/internal/world"
)

type Object struct {
	manifest      sprite.Sprite
	SpawnLocation world.Location
	Visible       bool
	Interactive   bool
	CanMove       bool
}
