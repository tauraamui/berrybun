package main

import (
	"flag"
	_ "image/png"

	"github.com/tacusci/logging"

	"github.com/hajimehoshi/ebiten"
	"github.com/tauraamui/berrybun/game"
)

func parseOptionFlags(g *game.Game) {
	flag.BoolVar(&g.Debug, "dbg", false, "Enable game's debug mode")
	flag.BoolVar(&g.Fullscreen, "fs", false, "Set game to be fullscreen")
	flag.BoolVar(&g.AllowKeyboard, "allowkbrd", false, "Turns on accepting keyboard based controls")

	flag.Parse()
}

func main() {
	var game = game.Game{}

	parseOptionFlags(&game)

	if game.Debug {
		logging.SetLevel(logging.DebugLevel)
	}

	game.Init()

	w, h := ebiten.MonitorSize()
	// On mobiles, ebiten.MonitorSize is not available so far.
	// Use arbitrary values.
	if w == 0 || h == 0 {
		w = 300
		h = 450
	}

	ebiten.SetFullscreen(game.Fullscreen)

	s := ebiten.DeviceScaleFactor()

	var sw, sh int = 800, 600

	if game.Fullscreen {
		sw, sh = int(float64(w)*s), int(float64(h)*s)
	}

	if err := ebiten.Run(game.Update, sw, sh, 1/s, "Berrybun Game"); err != nil {
		panic(err)
	}
}
