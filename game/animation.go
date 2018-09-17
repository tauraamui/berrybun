package game

import (
	"image"

	"github.com/hajimehoshi/ebiten"
)

type Animation struct {
	id                 uint
	spritesheet        *ebiten.Image
	repeatLoopStart    int
	repeatLoopEnd      int
	maxRepeatLoopCount int
	repeatLoopCount    int
	frameWidth         int
	frameHeight        int
	frame0X            int
	frame0Y            int
	frameNum           int
	defaultSpeed       int
	speed              int
	count              int
	countForLoopStart  int
}

func (a *Animation) Update(screen *ebiten.Image) error {

	if a.count < 0 {
		a.count = 0
	}

	a.count++

	scale := ebiten.DeviceScaleFactor()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(a.frameWidth)/2, -float64(a.frameHeight)/2)
	sw, sh := screen.Size()
	swf := float64(sw) - (float64(sw) * float64(0.9991))
	shf := float64(sh) - (float64(sh) * float64(0.9988))
	op.GeoM.Scale(scale+swf, scale+shf)
	op.GeoM.Translate(float64(sw)/2, float64(sh)/2)
	i := (a.count / a.speed) % a.frameNum

	//if current frame is now on start loop sprite
	if i == a.repeatLoopStart {
		//get the count value for this position
		a.countForLoopStart = a.count
	}

	//if current frame is now at end of loop
	if i == a.repeatLoopEnd {
		//if more loops to do
		if a.repeatLoopCount < a.maxRepeatLoopCount {
			//set animation back to the start of the loop
			a.count = a.countForLoopStart
			a.repeatLoopCount++
		}
	}

	//if maximum number of loops occurred
	if a.repeatLoopCount >= a.maxRepeatLoopCount {
		//if current frame is end of animation
		if i == a.frameNum-1 {
			//set the loop count back
			a.repeatLoopCount = 0
		}
	}

	sx, sy := a.frame0X+i*a.frameWidth, a.frame0Y
	r := image.Rect(sx, sy, sx+a.frameWidth, sy+a.frameHeight)
	op.SourceRect = &r

	var err error
	if !ebiten.IsDrawingSkipped() {
		err = screen.DrawImage(a.spritesheet, op)
	}

	if err != nil {
		return err
	}

	return nil
}

func (a *Animation) Reset() {
	a.count = -1
	a.speed = a.defaultSpeed
	a.repeatLoopCount = 0
}
