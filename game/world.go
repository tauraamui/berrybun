package game

import (
	"github.com/hajimehoshi/ebiten"
)

type World struct {
	game   *Game
	player *Player
}

func (w *World) Init() {
	w.player.Init()
}

func (w *World) Update(screen *ebiten.Image) error {
	w.player.Update(screen)
	return nil
}
