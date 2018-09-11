package game

import (
	"image"
	"os"

	"github.com/hajimehoshi/ebiten"
)

type World struct {
	game   *Game
	wMap   *Map
	player *Player
}

func (w *World) Init() {
	w.wMap = &Map{
		world:   w,
		bglayer: []uint32{23, 23},
	}
	w.player.Init()
}

func (w *World) Update(screen *ebiten.Image) error {
	w.player.Update(screen)
	return nil
}

type Map struct {
	world   *World
	bglayer []uint32
}

func (m *Map) Update(screen *ebiten.Image) error {
	return nil
}

type Player struct {
	game                *Game
	animation           *Animation
	idleAnimation       *Animation
	hopRightAnimation   *Animation
	hopLeftAnimation    *Animation
	hopForwardAnimation *Animation
	hopDownAnimation    *Animation
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
		id:           2,
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
		id:           3,
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
		id:           4,
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

	p.hopDownAnimation = &Animation{
		id:           5,
		spritesheet:  animSpriteSheet,
		frameWidth:   32,
		frameHeight:  32,
		frame0X:      0,
		frame0Y:      128,
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
			if p.animation.id != p.hopDownAnimation.id {
				p.animation.Reset()
				p.hopDownAnimation.Reset()
				p.animation = p.hopDownAnimation
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

		//vary speed relative to how much joystick pushed in either direction
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
