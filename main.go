package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 8
)

var (
	count  = 0
	sprite *ebiten.Image
)

func init() {

	imgFile, err := os.Open("./res/image.png")

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
	if ebiten.IsDrawingSkipped() {
		return nil
	}

	count++

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	op.GeoM.Translate(screenWidth/2, screenHeight/2)
	i := (count / 30) % frameNum
	sx, sy := frameOX+i*frameWidth, frameOY
	r := image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)
	op.SourceRect = &r
	screen.DrawImage(sprite, op)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS()))
	return nil
}

func main() {
	if err := ebiten.Run(update, screenWidth, screenHeight, 2, "Berrybun Game"); err != nil {
		panic(err)
	}
}
