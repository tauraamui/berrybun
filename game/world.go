package game

import (
	"github.com/hajimehoshi/ebiten"
)

type World struct {
	game   *Game
	wMap   *Map
	player *Player
}

func (w *World) Init() {
	w.wMap = &Map{
		world:   w,
		bglayer: []uint32{23, 23},
	}
	w.player.Init()
}

func (w *World) Update(screen *ebiten.Image) error {
	w.player.Update(screen)
	return nil
}

type Map struct {
	world   *World
	bglayer []uint32
}

func (m *Map) Update(screen *ebiten.Image) error {
	return nil
}
