package game

import (
	"fmt"
	"image"
	"math/rand"
	"os"
	"runtime"
	"time"

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
	world                     *World
	bgSpriteSheet             *ebiten.Image
	bglayer                   [][]int
	skippedTileLastOutputTime time.Time
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
		if y%6 == 0 {
			var grassOnRow = 0
			for i := 0; i < len(newRow); i++ {
				if i > 2 && rand.Intn(2) == 1 && grassOnRow < 30 {
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

	skippedTileCount := 0

	sw, sh := screen.Size()
	swf := float64(sw) - (float64(sw) * float64(0.9991))
	shf := float64(sh) - (float64(sh) * float64(0.9988))

	for y := 0; y < len(m.bglayer); y++ {
		xTiles := len(m.bglayer[y])
		for x := 0; x < xTiles; x++ {

			tileXPos, tileYPos := 0.0, 0.0

			tileXPos += float64((x % xTiles) * spriteSize)
			tileYPos += float64(y * spriteSize)

			var tileWidth, tileHeight float64 = 16, 16
			tileWidth *= scale + swf
			tileHeight *= scale + shf

			if int(tileXPos) > m.world.game.cameraX+sw {
				skippedTileCount++
				continue
			}

			if int(tileXPos) < m.world.game.cameraX && int(tileXPos+tileWidth) < m.world.game.cameraX {
				skippedTileCount++
				continue
			}

			if int(tileYPos) > (m.world.game.cameraY*-1)+sh {
				continue
			}

			if int(tileYPos+tileHeight) < m.world.game.cameraY*-1 {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((x%xTiles)*spriteSize), float64(y*spriteSize))
			op.GeoM.Translate(float64(m.world.game.cameraX*-1), float64(m.world.game.cameraY))
			op.GeoM.Scale(scale+swf, scale+shf)

			r := image.Rect(m.bglayer[y][x]*spriteSize, 0, (m.bglayer[y][x]+1)*spriteSize, (m.bglayer[y][x]+1)*spriteSize)
			op.SourceRect = &r

			if err := screen.DrawImage(m.bgSpriteSheet, op); err != nil {
				return err
			}
		}
	}

	if time.Since(m.skippedTileLastOutputTime) > time.Second*3 {
		fmt.Printf("Skipped %d tiles\n", skippedTileCount)
		m.skippedTileLastOutputTime = time.Now()
	}

	return nil
}

type Player struct {
	game                     *Game
	animation                *Animation
	idleAnimation            *Animation
	hopRightAnimation        *Animation
	hopLeftAnimation         *Animation
	hopForwardAnimation      *Animation
	hopDownAnimation         *Animation
	hopForwardLeftAnimation  *Animation
	hopForwardRightAnimation *Animation

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
		frame0Y:      160,
		frameNum:     6,
		defaultSpeed: 8,
		speed:        8,
		count:        -1,
	}

	p.hopForwardRightAnimation = &Animation{
		id:           6,
		spritesheet:  animSpriteSheet,
		frameWidth:   32,
		frameHeight:  32,
		frame0X:      0,
		frame0Y:      192,
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
		p.game.cameraY += 9 - p.animation.speed
	}

	if p.MovingDown() {
		p.game.cameraY -= 9 - p.animation.speed
	}

	if p.MovingRight() {
		p.game.cameraX += 9 - p.animation.speed
	}

	if p.MovingLeft() {
		p.game.cameraX -= 9 - p.animation.speed
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

		if p.MovingUpMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
	}

	if p.MovingDown() && !p.MovingRight() && !p.MovingLeft() {
		playerMoving = true
		if p.animation.id != p.hopDownAnimation.id {
			p.animation.Reset()
			p.hopDownAnimation.Reset()
			p.animation = p.hopDownAnimation
		}

		if p.MovingDownMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
	}

	if p.MovingRight() && !p.MovingUp() && !p.MovingDown() {
		playerMoving = true
		if p.animation.id != p.hopRightAnimation.id {
			p.animation.Reset()
			p.hopRightAnimation.Reset()
			p.animation = p.hopRightAnimation
		}

		if p.MovingRightMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
	}

	if p.MovingLeft() && !p.MovingUp() && !p.MovingDown() {
		playerMoving = true
		if p.animation.id != p.hopLeftAnimation.id {
			p.animation.Reset()
			p.hopLeftAnimation.Reset()
			p.animation = p.hopLeftAnimation
		}

		if p.MovingLeftMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
	}

	if p.MovingLeft() && p.MovingUp() {
		playerMoving = true
		if p.animation.id != p.hopForwardLeftAnimation.id {
			p.animation.Reset()
			p.hopLeftAnimation.Reset()
			p.animation = p.hopForwardLeftAnimation
		}

		if p.MovingLeftMore() || p.MovingUpMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
	}

	if p.MovingLeft() && p.MovingDown() {
	}

	if p.MovingRight() && p.MovingUp() {
		playerMoving = true
		if p.animation.id != p.hopForwardRightAnimation.id {
			p.animation.Reset()
			p.hopLeftAnimation.Reset()
			p.animation = p.hopForwardRightAnimation
		}

		if p.MovingRightMore() || p.MovingUpMore() {
			if p.animation.speed > 6 {
				p.animation.speed--
			}
		} else {
			p.animation.speed = p.animation.defaultSpeed
		}
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
	if len(p.game.gamepads) > 0 {
		return p.game.gamepads[0].axes[0] >= 0.30
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyD)
	}
	return false
}

func (p *Player) MovingRightMore() bool {
	if len(p.game.gamepads) > 0 {
		return p.game.gamepads[0].axes[0] >= 0.80
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyD)
	}
	return false
}

func (p *Player) MovingLeft() bool {
	if len(p.game.gamepads) > 0 {
		return p.game.gamepads[0].axes[0] <= -0.30
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyA)
	}
	return false
}

func (p *Player) MovingLeftMore() bool {
	if len(p.game.gamepads) > 0 {
		return p.game.gamepads[0].axes[0] <= -0.80
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyA)
	}
	return false
}

func (p *Player) MovingUp() bool {
	if len(p.game.gamepads) > 0 {
		if runtime.GOOS != "windows" {
			return p.game.gamepads[0].axes[1] <= -0.30
		}
		return p.game.gamepads[0].axes[1] >= 0.30
	}

	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyW)
	}
	return false
}

func (p *Player) MovingUpMore() bool {
	if len(p.game.gamepads) > 0 {
		if runtime.GOOS != "windows" {
			return p.game.gamepads[0].axes[1] <= -0.80
		}
		return p.game.gamepads[0].axes[1] >= 0.80
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyW)
	}
	return false
}

func (p *Player) MovingDown() bool {
	if len(p.game.gamepads) > 0 {
		if runtime.GOOS != "windows" {
			return p.game.gamepads[0].axes[1] >= 0.30
		}
		return p.game.gamepads[0].axes[1] <= -0.30
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyS)
	}
	return false
}

func (p *Player) MovingDownMore() bool {
	if len(p.game.gamepads) > 0 {
		if runtime.GOOS != "windows" {
			return p.game.gamepads[0].axes[1] >= 0.80
		}
		return p.game.gamepads[0].axes[1] <= -0.80
	}
	if p.game.AllowKeyboard {
		return ebiten.IsKeyPressed(ebiten.KeyS)
	}

	return false
}
