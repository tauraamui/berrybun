package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten"
	"github.com/tauraamui/berrybun/game"
)

func main() {
	var game = game.Game{}
	game.Init()

	//ebiten.SetFullscreen(true)

	w, h := ebiten.MonitorSize()
	// On mobiles, ebiten.MonitorSize is not available so far.
	// Use arbitrary values.
	if w == 0 || h == 0 {
		w = 300
		h = 450
	}

	// ebiten.SetFullscreen(true)

	s := ebiten.DeviceScaleFactor()

	// if err := ebiten.Run(game.Update, int(float64(w)*s), int(float64(h)*s), 1/s, "Berrybun Game"); err != nil {
	if err := ebiten.Run(game.Update, 800, 600, 1/s, "Berrybun Game"); err != nil {
		panic(err)
	}
}
