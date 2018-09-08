package core

import (
	"image"
	"log"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten"
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
	mu       sync.Mutex
	player   *Player
	gamepads []GamePadInput
}

func (g *Game) Init() {
	g.player = &Player{
		game: g,
	}
	g.player.Init()
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
		log.Printf("gamepad connected: id: %d", id)
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
			log.Printf("gamepad disconnected: id: %d", g.gamepads[i].id)
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

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	return g.player.Update(screen)

}

type Player struct {
	game                *Game
	animation           *Animation
	idleAnimation       *Animation
	hopRightAnimation   *Animation
	hopLeftAnimation    *Animation
	hopForwardAnimation *Animation
}

func (p *Player) Init() {

	bunnyAnimationsFile, err := os.Open("./res/bunny.png")

	if err != nil {
		panic(err)
	}

	defer bunnyAnimationsFile.Close()

	img, _, err := image.Decode(bunnyAnimationsFile)

	if err != nil {
		panic(err)
	}

	animSpriteSheet, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	if err != nil {
		panic(err)
	}

	p.idleAnimation = &Animation{
		id:                 0,
		spritesheet:        animSpriteSheet,
		repeatLoopStart:    0,
		repeatLoopEnd:      1,
		maxRepeatLoopCount: 200,
		frameWidth:         32,
		frameHeight:        32,
		frame0X:            0,
		frame0Y:            0,
		frameNum:           6,
		defaultSpeed:       2,
		speed:              2,
		count:              -1,
	}

	p.hopRightAnimation = &Animation{
		id:           1,
		spritesheet:  animSpriteSheet,
		frameWidth:   32,
		frameHeight:  32,
		frame0X:      0,
		frame0Y:      32,
		frameNum:     6,
		defaultSpeed: 8,
		speed:        8,
		count:        -1,
	}

	p.hopLeftAnimation = &Animation{
		id:           1,
		spritesheet:  animSpriteSheet,
		frameWidth:   32,
		frameHeight:  32,
		frame0X:      0,
		frame0Y:      64,
		frameNum:     6,
		defaultSpeed: 8,
		speed:        8,
		count:        -1,
	}

	p.hopForwardAnimation = &Animation{
		id:           1,
		spritesheet:  animSpriteSheet,
		frameWidth:   32,
		frameHeight:  32,
		frame0X:      0,
		frame0Y:      96,
		frameNum:     6,
		defaultSpeed: 8,
		speed:        8,
		count:        -1,
	}

	p.animation = p.idleAnimation
}

func (p *Player) Update(screen *ebiten.Image) error {

	if len(p.game.gamepads) > 0 {
		j1LeftRightAxes := p.game.gamepads[0].axes[0]
		j1UpDownAxes := p.game.gamepads[0].axes[1]

		playerMoving := false

		if j1UpDownAxes >= 0.30 {
			playerMoving = true
			if p.animation.id != p.hopForwardAnimation.id {
				p.animation.Reset()
				p.hopForwardAnimation.Reset()
				p.animation = p.hopForwardAnimation
			}
		} else if j1UpDownAxes <= -0.30 {
			playerMoving = true
			if p.animation.id != p.hopForwardAnimation.id {
				p.animation.Reset()
				p.hopForwardAnimation.Reset()
				p.animation = p.hopForwardAnimation
			}
		}

		if j1LeftRightAxes >= 0.30 {
			playerMoving = true
			//force previous/existing animation loop to reset to 0
			if p.animation.id != p.hopRightAnimation.id {
				p.animation.Reset()
				p.hopRightAnimation.Reset()
				p.animation = p.hopRightAnimation
			}
		} else if j1LeftRightAxes <= -0.30 {
			playerMoving = true
			if p.animation.id != p.hopLeftAnimation.id {
				p.animation.Reset()
				p.hopLeftAnimation.Reset()
				p.animation = p.hopLeftAnimation
			}
		}

		if j1LeftRightAxes >= 0.80 || j1LeftRightAxes <= -0.80 {
			if p.animation.speed > 5 {
				p.animation.speed--
			}
		} else if j1UpDownAxes >= 0.80 || j1UpDownAxes <= -0.80 {
			if p.animation.speed > 5 {
				p.animation.speed--
			}
		} else {
			if p.animation.speed < 8 {
				p.animation.speed++
			}
		}

		if !playerMoving {
			if p.animation.id != p.idleAnimation.id {
				p.animation.Reset()
				p.idleAnimation.Reset()
				p.animation = p.idleAnimation
			}
		}
	}

	p.animation.Update(screen)

	return nil
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
	swf := float64(sw) - (float64(sw) * float64(0.9988))
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
	err := screen.DrawImage(a.spritesheet, op)

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
