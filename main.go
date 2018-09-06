package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/tauraamui/berrybun/core"
)

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX = 0
	// frameOY = 0
	// frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 6
)

var (
	count          = 0
	sprite         *ebiten.Image
	gamepadIDs     = map[int]struct{}{}
	frameOY        = 0
	animationSpeed = 8
)

func init() {

	imgFile, err := os.Open("./res/bunny.png")

	if err != nil {
		panic(err)
	}

	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)

	if err != nil {
		panic(err)
	}

	sprite, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}

func update(screen *ebiten.Image) error {

	// Log the gamepad connection events.
	for _, id := range inpututil.JustConnectedGamepadIDs() {
		log.Printf("gamepad connected: id: %d", id)
		gamepadIDs[id] = struct{}{}
	}
	for id := range gamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			log.Printf("gamepad disconnected: id: %d", id)
			delete(gamepadIDs, id)
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

			// Log button events.
			if inpututil.IsGamepadButtonJustPressed(id, b) {
				log.Printf("button pressed: id: %d, button: %d", id, b)
			}
			if inpututil.IsGamepadButtonJustReleased(id, b) {
				log.Printf("button released: id: %d, button: %d", id, b)
			}
		}
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	count++

	op := &ebiten.DrawImageOptions{}
	//move sprite backwards and up by half of its width and height
	op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	//move sprite's origin to half of screen in width and height
	op.GeoM.Translate(screenWidth/2, screenHeight/2)
	//speed of changing from one animation frame to another
	i := (count / animationSpeed) % frameNum
	sx, sy := frameOX+i*frameWidth, frameOY
	r := image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)
	op.SourceRect = &r
	screen.DrawImage(sprite, op)

	// Draw the current gamepad status.
	str := ""
	str += fmt.Sprintf("FPS: %0.2f\n", ebiten.CurrentFPS())

	if axes[0] >= 0.5 {
		frameOY = 32
		if axes[0] >= 0.9 {
			if animationSpeed > 5 {
				animationSpeed--
			} else {
				animationSpeed = 5
			}
		} else {
			animationSpeed = 8
		}
	} else {
		frameOY = 0
		animationSpeed = 8
	}

	if len(ids) > 0 {

	} else {
		str = "Please connect your gamepad."
	}
	ebitenutil.DebugPrint(screen, str)

	return nil
}

func main() {
	var game = core.Game{}
	game.Init()

	if err := ebiten.Run(game.Update, screenWidth, screenHeight, 2, "Berrybun Game"); err != nil {
		panic(err)
	}
}
