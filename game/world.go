package game

import (
	"image"
	"math/rand"
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

	m.bglayer = make([][]int, 40)

	for y := 0; y < len(m.bglayer); y++ {
		newRow := make([]int, 50)
		if y%6 == 0 {
			var grassOnRow = 0
			for i := 0; i < len(newRow); i++ {
				if i > 2 && rand.Intn(3) == 2 && grassOnRow < 5 {
					grass := rand.Intn(2)
					newRow[i] = grass
					if grass == 1 {
						grassOnRow++
					}
				}
			}
		}
		m.bglayer[y] = newRow
	}

	return nil
}

func (m *Map) Update(screen *ebiten.Image) error {

	const (
		spriteSize = 16
	)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	scale := ebiten.DeviceScaleFactor()

	for y := 0; y < len(m.bglayer); y++ {
		xTiles := len(m.bglayer[y])
		for x := 0; x < xTiles; x++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((x%xTiles)*spriteSize), float64(y*spriteSize))
			op.GeoM.Translate(float64(m.world.game.cameraX*-1), float64(m.world.game.cameraY))
			sw, sh := screen.Size()
			swf := float64(sw) - (float64(sw) * float64(0.9991))
			shf := float64(sh) - (float64(sh) * float64(0.9988))
			op.GeoM.Scale(scale+swf, scale+shf)
			r := image.Rect(m.bglayer[y][x]*spriteSize, 0, (m.bglayer[y][x]+1)*spriteSize, (m.bglayer[y][x]+1)*spriteSize)
			op.SourceRect = &r

			if err := screen.DrawImage(m.bgSpriteSheet, op); err != nil {
				return err
			}
		}
	}

	return nil
}

type Player struct {
	game                    *Game
	animation               *Animation
	idleAnimation           *Animation
	hopRightAnimation       *Animation
	hopLeftAnimation        *Animation
	hopForwardAnimation     *Animation
	hopDownAnimation        *Animation
	hopForwardLeftAnimation *Animation

	speed int
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

	p.hopForwardLeftAnimation = &Animation{
		id:           6,
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

	p.animation = p.idleAnimation
}

func (p *Player) Update(screen *ebiten.Image) error {

	p.Move()
	p.animation.Update(screen)

	return nil
}

func (p *Player) Move() {

	if p.MovingUp() {
		p.game.cameraY++
	}

	if p.MovingDown() {
		p.game.cameraY--
	}

	if p.MovingRight() {
		p.game.cameraX++
	}

	if p.MovingLeft() {
		p.game.cameraX--
	}

	p.UpdateAnimation()
}

func (p *Player) UpdateAnimation() {
	playerMoving := false
	if p.MovingUp() && !p.MovingRight() && !p.MovingLeft() {
		playerMoving = true
		if p.animation.id != p.hopForwardAnimation.id {
			p.animation.Reset()
			p.hopForwardAnimation.Reset()
			p.animation = p.hopForwardAnimation
		}
	}

	if p.MovingDown() && !p.MovingRight() && !p.MovingLeft() {
		playerMoving = true
		if p.animation.id != p.hopDownAnimation.id {
			p.animation.Reset()
			p.hopDownAnimation.Reset()
			p.animation = p.hopDownAnimation
		}
	}

	if p.MovingRight() && !p.MovingUp() && !p.MovingDown() {
		playerMoving = true
		if p.animation.id != p.hopRightAnimation.id {
			p.animation.Reset()
			p.hopRightAnimation.Reset()
			p.animation = p.hopRightAnimation
		}
	}

	if p.MovingLeft() && !p.MovingUp() && !p.MovingDown() {
		playerMoving = true
		if p.animation.id != p.hopLeftAnimation.id {
			p.animation.Reset()
			p.hopLeftAnimation.Reset()
			p.animation = p.hopLeftAnimation
		}
	}

	if p.MovingLeft() && p.MovingUp() {
		playerMoving = true
		if p.animation.id != p.hopForwardLeftAnimation.id {
			p.animation.Reset()
			p.hopLeftAnimation.Reset()
			p.animation = p.hopLeftAnimation
		}
	}

	if p.MovingLeft() && p.MovingDown() {
	}

	if p.MovingRight() && p.MovingUp() {
	}

	if p.MovingRight() && p.MovingDown() {
	}

	if !playerMoving {
		if p.animation.id != p.idleAnimation.id {
			p.animation.Reset()
			p.idleAnimation.Reset()
			p.animation = p.idleAnimation
		}
	}
}

func (p *Player) MovingRight() bool {
	if len(p.game.gamepads) == 0 {
		return false
	}
	return p.game.gamepads[0].axes[0] >= 0.30
}

func (p *Player) MovingLeft() bool {
	if len(p.game.gamepads) == 0 {
		return false
	}
	return p.game.gamepads[0].axes[0] <= -0.30
}

func (p *Player) MovingUp() bool {
	if len(p.game.gamepads) == 0 {
		return false
	}
	return p.game.gamepads[0].axes[1] >= 0.30
}

func (p *Player) MovingDown() bool {
	if len(p.game.gamepads) == 0 {
		return false
	}
	return p.game.gamepads[0].axes[1] <= -0.30
}
