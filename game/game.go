package game

import (
	"fmt"
	"image"
	"sync"

	"github.com/tacusci/logging"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

type GamePadInput struct {
	id   int
	axes []float64
}

func (gpi *GamePadInput) update() {
	for a := 0; a < len(gpi.axes); a++ {
		v := ebiten.GamepadAxis(gpi.id, a)
		gpi.axes[a] = v
	}
}

type Game struct {
	mu            sync.Mutex
	Debug         bool
	AllowKeyboard bool
	cameraX       int
	cameraY       int
	world         *World
	gamepads      []GamePadInput
}

func (g *Game) Init() {
	g.cameraX = 0
	g.cameraY = 0
	g.world = &World{
		game: g,
		player: &Player{
			game: g,
		},
	}
	g.world.Init()
}

//AddGamepad adds a gamepad struct to collection if doesn't already contain gamepad of same id
func (g *Game) AddGamepad(gp GamePadInput) {
	g.mu.Lock()
	defer g.mu.Unlock()
	gamepadAlreadyInList := false
	for _, existingGamepad := range g.gamepads {
		if existingGamepad.id == gp.id {
			gamepadAlreadyInList = true
			break
		}
	}
	if !gamepadAlreadyInList {
		g.gamepads = append(g.gamepads, gp)
	}
}

//DeleteGamepad finds gamepad of provided id and then deletes from collection
func (g *Game) DeleteGamepad(gpid int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	deleteIndex := -1
	for i := 0; i < len(g.gamepads); i++ {
		if g.gamepads[i].id == gpid {
			deleteIndex = i
			break
		}
	}
	if deleteIndex > 0 {
		g.gamepads = append(g.gamepads[:deleteIndex], g.gamepads[:deleteIndex+1]...)
	}
}

func (g *Game) updateGamepads() {
	// check for any disconnected gamepads and remove from game
	for _, id := range inpututil.JustConnectedGamepadIDs() {
		if logging.CurrentLoggingLevel == logging.DebugLevel {
			logging.Debug(fmt.Sprintf("gamepad connected: id: %d", id))
		}
		gamepadAlreadyInList := false
		for _, gp := range g.gamepads {
			if gp.id == id {
				gamepadAlreadyInList = true
				break
			}
		}
		if !gamepadAlreadyInList {
			g.AddGamepad(GamePadInput{
				id:   id,
				axes: make([]float64, ebiten.GamepadAxisNum(id)),
			})
		}
	}

	// check for any connected gamepads and add them to the game
	for i := 0; i < len(g.gamepads); i++ {
		if inpututil.IsGamepadJustDisconnected(g.gamepads[i].id) {
			if logging.CurrentLoggingLevel == logging.DebugLevel {
				logging.Debug(fmt.Sprintf("gamepad disconnected: id: %d", g.gamepads[i].id))
			}
			g.DeleteGamepad(g.gamepads[i].id)
		}
	}

	for i := 0; i < len(g.gamepads); i++ {
		g.gamepads[i].update()
	}
}

//Update updates everything within game state
func (g *Game) Update(screen *ebiten.Image) error {
	g.updateGamepads()
	if err := g.world.Update(screen); err != nil {
		return err
	}

	if err := ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())); err != nil {
		return nil
	}

	return nil

	// if ebiten.IsDrawingSkipped() {
	// 	return nil
	// }
}

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
