package core

import (
	"image"
	"log"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type Game struct {
	player     *Player
	gamepadIDs map[int]struct{}
}

func (g *Game) Init() {
	g.player = &Player{
		animSpriteSheetLocation: "./res/bunny.png",
	}
	g.player.Init()
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.updateGamepad()
	return nil
}

func (g *Game) updateGamepad() {
	// Log the gamepad connection events.
	for _, id := range inpututil.JustConnectedGamepadIDs() {
		log.Printf("gamepad connected: id: %d", id)
		g.gamepadIDs[id] = struct{}{}
	}
	for id := range g.gamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			log.Printf("gamepad disconnected: id: %d", id)
			delete(g.gamepadIDs, id)
		}
	}

	ids := ebiten.GamepadIDs()
	axes := []float64{}
	pressedButtons := map[int][]string{}

	for _, id := range ids {
		maxAxis := ebiten.GamepadAxisNum(id)
		for a := 0; a < maxAxis; a++ {
			v := ebiten.GamepadAxis(id, a)
			axes = append(axes, v)
		}
		maxButton := ebiten.GamepadButton(ebiten.GamepadButtonNum(id))
		for b := ebiten.GamepadButton(id); b < maxButton; b++ {
			if ebiten.IsGamepadButtonPressed(id, b) {
				pressedButtons[id] = append(pressedButtons[id], strconv.Itoa(int(b)))
			}
		}
	}
}

type Player struct {
	animSpriteSheetLocation string
	animSpriteSheet         *ebiten.Image
}

func (p *Player) Init() {
	imgFile, err := os.Open(p.animSpriteSheetLocation)

	if err != nil {
		panic(err)
	}

	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)

	if err != nil {
		panic(err)
	}

	p.animSpriteSheet, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}
