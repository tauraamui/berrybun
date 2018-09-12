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
		world: w,
	}
	w.wMap.Init()
	w.player.Init()
}

func (w *World) Update(screen *ebiten.Image) error {
	w.wMap.Update(screen)
	w.player.Update(screen)
	return nil
}

type Map struct {
	world         *World
	bgSpriteSheet *ebiten.Image
	bglayer       [][]int
}

func (m *Map) Init() error {
	mapTileSizeShape, err := os.Open("./res/map.png")

	if err != nil {
		panic(err)
	}

	defer mapTileSizeShape.Close()

	img, _, err := image.Decode(mapTileSizeShape)

	if err != nil {
		panic(err)
	}

	m.bgSpriteSheet, err = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	if err != nil {
		panic(err)
	}

	m.bglayer = make([][]int, 500)

	for y := 0; y < len(m.bglayer); y++ {
		newRow := make([]int, 500)
		if y%5 == 0 {
			for i := 0; i < len(newRow); i++ {
				newRow[i] = 1
			}
		}
		m.bglayer[y] = newRow
	}

	return nil
}

func (m *Map) Update(screen *ebiten.Image) error {

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	for y := 0; y < len(m.bglayer); y++ {
		xTiles := len(m.bglayer[y])
		for x := 0; x < xTiles; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((x%xTiles)*16), float64(y*16))
			r := image.Rect(m.bglayer[y][x]*16, m.bglayer[y][x]*16, (m.bglayer[y][x]+1)*16, (m.bglayer[y][x]+1)*16)
			op.SourceRect = &r

			if err := screen.DrawImage(m.bgSpriteSheet, op); err != nil {
				return err
			}
		}
	}

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
			if p.animation.id != p.hopForwardAnimation.id {
				p.animation.Reset()
				p.hopForwardAnimation.Reset()
				p.animation = p.hopForwardAnimation
			}
		} else if j1UpDownAxes <= -0.30 {
			playerMoving = true
			if p.animation.id != p.hopDownAnimation.id {
				p.animation.Reset()
				p.hopDownAnimation.Reset()
				p.animation = p.hopDownAnimation
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
