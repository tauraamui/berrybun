package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten"
	"github.com/tauraamui/berrybun/core"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

func main() {
	var game = core.Game{}
	game.Init()

	if err := ebiten.Run(game.Update, screenWidth, screenHeight, 2, "Berrybun Game"); err != nil {
		panic(err)
	}
}
