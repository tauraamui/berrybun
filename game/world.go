package game

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/tacusci/logging"
	"github.com/tauraamui/berrybun/res"
	"github.com/tauraamui/berrybun/utils"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/images"
)

type World struct {
	game           *Game
	wMap           *Map
	player         *Player
	nightTime      bool
	spotLightImage *ebiten.Image
	maskedFgImage  *ebiten.Image
	fgImage        *ebiten.Image
}

func (w *World) Init() {
	w.wMap = &Map{
		game:     w.game,
		bgwidth:  150,
		bgheight: 150,
	}
	w.wMap.Init()
	w.player.Init()

	// Initialize the spot light image.
	const r = 64
	alphas := image.Point{r * 2, r * 2}
	a := image.NewAlpha(image.Rectangle{image.ZP, alphas})
	for j := 0; j < alphas.Y; j++ {
		for i := 0; i < alphas.X; i++ {
			// d is the distance between (i, j) and the (circle) center.
			d := math.Sqrt(float64((i-r)*(i-r) + (j-r)*(j-r)))
			// Alphas around the center are 0 and values outside of the circle are 0xff.
			b := uint8(utils.Max(0, utils.Min(0xff, int(3*d*0xff/r)-2*0xff)))
			a.SetAlpha(i, j, color.Alpha{b})
		}
	}
	w.spotLightImage, _ = ebiten.NewImageFromImage(a, ebiten.FilterDefault)

	w.maskedFgImage, _ = ebiten.NewImage(screenWidth, screenHeight, ebiten.FilterDefault)

	img, _, err := image.Decode(bytes.NewReader(images.FiveYears_jpg))
	if err != nil {
		log.Fatal(err)
	}
	w.fgImage, _ = ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}

func (w *World) Update(screen *ebiten.Image) error {
	w.wMap.Update(screen)
	w.player.Update(screen)
	return nil
}

type Map struct {
	game                      *Game
	bgSpriteSheet             *ebiten.Image
	bglayer                   [][]int
	bgwidth                   int
	bgheight                  int
	buildings                 []Building
	skippedTileLastOutputTime time.Time
}

func (m *Map) Init() error {

	img, _, err := image.Decode(bytes.NewReader(res.Map_png))

	if err != nil {
		log.Fatal(err)
	}

	m.bgSpriteSheet, err = ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	if err != nil {
		panic(err)
	}

	m.bglayer = make([][]int, m.bgheight)

	// initial values of slices will be zero, so set random indexes to be grass or flowers or something else
	for y := 0; y < len(m.bglayer); y++ {
		newRow := make([]int, m.bgwidth)
		for x := 0; x < len(m.bglayer); x++ {
			newRow[x] = utils.CombineNumbers(float64(18), float64(9))
			if y%(rand.Intn(4)+1) == 0 {
				var grassOnRow = 0
				if x > 2 && rand.Intn(2) == 1 && grassOnRow < int(float64(m.bgwidth)*0.75) {
					grass := rand.Intn(3)
					if grass == 1 {
						newRow[x] = utils.CombineNumbers(float64(18), float64(8))
					} else if grass == 2 {
						newRow[x] = utils.CombineNumbers(float64(19), float64(8))
					}
					if grass == 1 || grass == 2 {
						grassOnRow++
					}
				}
			}
		}
		m.bglayer[y] = newRow
	}

	m.buildings = append(m.buildings, Building{
		game:        m.game,
		spritesheet: m.bgSpriteSheet,
		x:           15,
		y:           15,
		width:       7,
		height:      7,
		tileXY:      utils.CombineNumbers(float64(1), float64(1)),
	})

	m.buildings = append(m.buildings, Building{
		game:        m.game,
		spritesheet: m.bgSpriteSheet,
		x:           239,
		y:           15,
		width:       7,
		height:      7,
		tileXY:      utils.CombineNumbers(float64(1), float64(1)),
	})

	m.buildings = append(m.buildings, Building{
		game:        m.game,
		spritesheet: m.bgSpriteSheet,
		x:           463,
		y:           15,
		width:       7,
		height:      7,
		tileXY:      utils.CombineNumbers(float64(1), float64(1)),
	})

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
	// work out width/height scale factor based on percentage of screen size
	swf := float64(sw / m.game.cameraWidth)
	shf := float64(sh / m.game.cameraHeight)

	// for each y row of map tiles
	for y := 0; y < len(m.bglayer); y++ {
		xTiles := len(m.bglayer[y])
		// for each tile in row
		for x := 0; x < xTiles; x++ {

			// set zeroed tile position
			tileXPos, tileYPos := 0.0, 0.0

			// transform tile positon based on tile matrix index
			tileXPos += float64((x % xTiles) * spriteSize)
			tileYPos += float64(y * spriteSize)

			// calculate tile width/height after scaling
			var tileWidth, tileHeight float64 = 16, 16
			tileWidth *= scale * swf
			tileHeight *= scale * shf

			// put the least intensive and most indicitive bound checks first for speed purposes

			// if the tile's x axis pos is further than the right edge of the screen, skip rendering tile
			if int(tileXPos) > m.game.cameraX+sw {
				skippedTileCount++
				continue
			}

			// if the tile's x axis pos is between the left and right edges of the screen, skip rendering tile
			if int(tileXPos) < m.game.cameraX && int(tileXPos+tileWidth) < m.game.cameraX {
				skippedTileCount++
				continue
			}

			// if the tile's y position is further than the bottom edge of the screen, (the *-1 is to invert the camera Y pos) skip rendering tile
			if int(tileYPos) > (m.game.cameraY*-1)+sh {
				continue
			}

			// if the tile is completely out above the top edge of the screen, skip rendering tile
			if int(tileYPos+tileHeight) < m.game.cameraY*-1 {
				continue
			}

			if !ebiten.IsDrawingSkipped() {
				// set rendering location on screen
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64((x%xTiles)*spriteSize), float64(y*spriteSize))
				op.GeoM.Translate(float64(m.game.cameraX*-1), float64(m.game.cameraY))
				op.GeoM.Scale(scale*swf, scale*shf)

				// crop/select sprite from the spritesheet
				tileX, tileY := utils.SplitNumbers(m.bglayer[y][x])

				r := image.Rect(tileX*spriteSize, tileY*spriteSize, (tileX*spriteSize)+spriteSize, (tileY*spriteSize)+spriteSize)

				// r := image.Rect(m.bglayer[y][x]*spriteSize, 0, (m.bglayer[y][x]+1)*spriteSize, (m.bglayer[y][x]+1)*spriteSize)
				op.SourceRect = &r

				if m.game.world.nightTime {
					op.ColorM.ChangeHSV(0.0, 1.0, 0.4)
				}

				if err := screen.DrawImage(m.bgSpriteSheet, op); err != nil {
					return err
				}
			}
		}
	}

	if logging.CurrentLoggingLevel == logging.DebugLevel {
		if time.Since(m.skippedTileLastOutputTime) > time.Second*3 {
			logging.Debug(fmt.Sprintf("Skipped %d tiles", skippedTileCount))
			m.skippedTileLastOutputTime = time.Now()
		}
	}

	for i := 0; i < len(m.buildings); i++ {
		m.buildings[i].Update(screen)
	}

	return nil
}

//Player data about player instance
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

//Init initialise player's animations, load spritesheet etc.,
func (p *Player) Init() {

	img, _, err := image.Decode(bytes.NewReader(res.Bunny_png))

	if err != nil {
		log.Fatal(err)
	}

	animSpriteSheet, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)

	if err != nil {
		panic(err)
	}

	p.idleAnimation = &Animation{
		game:               p.game,
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
		game:         p.game,
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
		game:         p.game,
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
		game:         p.game,
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
		game:         p.game,
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
		game:         p.game,
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
		game:         p.game,
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

type Building struct {
	game        *Game
	spritesheet *ebiten.Image
	x           int
	y           int
	width       int
	height      int
	tileXY      int
}

func (b *Building) Update(screen *ebiten.Image) error {

	const (
		spriteSize = 16
	)

	if !ebiten.IsDrawingSkipped() {
		scale := ebiten.DeviceScaleFactor()

		sw, sh := screen.Size()
		// work out width/height scale factor based on percentage of screen size
		swf := float64(sw / b.game.cameraWidth)
		shf := float64(sh / b.game.cameraHeight)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(b.x+(b.width*spriteSize)), float64(b.y+(b.height*spriteSize)))
		op.GeoM.Translate(float64(b.game.cameraX*-1)/2, float64(b.game.cameraY)/2)
		op.GeoM.Scale(scale*swf, scale*shf)
		op.GeoM.Scale(2, 2)

		// crop/select sprite from the spritesheet
		tileX, tileY := utils.SplitNumbers(b.tileXY)

		r := image.Rect(tileX, tileY, tileX+(spriteSize*b.width), tileY+(spriteSize*b.height))

		op.SourceRect = &r

		if b.game.world.nightTime {
			op.ColorM.ChangeHSV(0.0, 1.0, 0.4)
		}

		if err := screen.DrawImage(b.spritesheet, op); err != nil {
			return err
		}
	}

	return nil
}
