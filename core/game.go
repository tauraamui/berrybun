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

var count = 0

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
		animSpriteSheetLocation: "./res/bunny.png",
		frameOX:                 0,
		frameOY:                 0,
		frameWidth:              32,
		frameHeight:             32,
		frameNum:                6,
		animationSpeed:          8,
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
	game                    *Game
	animSpriteSheetLocation string
	animSpriteSheet         *ebiten.Image
	frameOX                 int
	frameOY                 int
	frameWidth              int
	frameHeight             int
	frameNum                int
	animationSpeed          float32
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

func (p *Player) Update(screen *ebiten.Image) error {
	count++

	op := &ebiten.DrawImageOptions{}
	//move sprite backwards and up by half of its width and height
	op.GeoM.Translate(-float64(p.frameWidth)/2, -float64(p.frameHeight)/2)
	//move sprite's origin to half of screen in width and height
	op.GeoM.Translate(screenWidth/2, screenHeight/2)
	//speed of changing from one animation frame to another
	i := (count / int(p.animationSpeed)) % p.frameNum
	sx, sy := p.frameOX+i*p.frameWidth, p.frameOY
	r := image.Rect(sx, sy, sx+p.frameWidth, sy+p.frameHeight)
	op.SourceRect = &r
	err := screen.DrawImage(p.animSpriteSheet, op)

	if err != nil {
		return err
	}

	if len(p.game.gamepads) > 0 {
		joystick1 := p.game.gamepads[0].axes[0]
		if joystick1 >= 0.30 {
			p.frameOY = 32
			if joystick1 >= 0.80 {
				if p.animationSpeed > 5 {
					p.animationSpeed -= 0.01
				}
			} else {
				if p.animationSpeed < 8 {
					p.animationSpeed += 0.01
				}
			}
		} else {
			p.frameOY = 0
			p.animationSpeed = 8
		}
	}

	return nil
}

type Animation struct {
	speed int
}
