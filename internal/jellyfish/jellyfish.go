// Package jellyfish renders drifting ASCII jellyfish and their trailing tentacles.
package jellyfish

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	minWidth  = 60
	minHeight = 24
)

var (
	waterPalette = []string{
		"\x1b[38;5;17m",
		"\x1b[38;5;18m",
		"\x1b[38;5;19m",
		"\x1b[38;5;24m",
	}
	bubblePalette = []string{
		"\x1b[38;5;31m",
		"\x1b[38;5;38m",
		"\x1b[38;5;81m",
		"\x1b[38;5;153m",
	}
	cyanBell = []string{
		"\x1b[38;5;37m",
		"\x1b[38;5;44m",
		"\x1b[38;5;51m",
		"\x1b[38;5;159m",
		"\x1b[38;5;195m",
	}
	pinkBell = []string{
		"\x1b[38;5;55m",
		"\x1b[38;5;91m",
		"\x1b[38;5;127m",
		"\x1b[38;5;171m",
		"\x1b[38;5;225m",
	}
	blueBell = []string{
		"\x1b[38;5;24m",
		"\x1b[38;5;31m",
		"\x1b[38;5;67m",
		"\x1b[38;5;111m",
		"\x1b[38;5;189m",
	}
)

// Config controls the jellyfish animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a preset suited to a typical terminal window.
func DefaultConfig() Config {
	return Config{
		Width:      104,
		Height:     38,
		FrameDelay: 55 * time.Millisecond,
	}
}

func (c Config) normalize() Config {
	if c.Width < minWidth {
		c.Width = minWidth
	}
	if c.Height < minHeight {
		c.Height = minHeight
	}
	if c.FrameDelay <= 0 {
		c.FrameDelay = 55 * time.Millisecond
	}
	return c
}

type cell struct {
	glyph byte
	color string
	level int
}

type grid struct {
	width  int
	height int
	cells  [][]cell
}

func newGrid(width, height int) *grid {
	g := &grid{
		width:  width,
		height: height,
		cells:  make([][]cell, height),
	}
	for y := range g.cells {
		g.cells[y] = make([]cell, width)
	}
	g.clear()
	return g
}

func (g *grid) clear() {
	for y := range g.cells {
		for x := range g.cells[y] {
			g.cells[y][x] = cell{glyph: ' ', level: -1}
		}
	}
}

func (g *grid) set(x, y int, glyph byte, color string, level int) {
	if x < 0 || x >= g.width || y < 0 || y >= g.height {
		return
	}
	if level < g.cells[y][x].level {
		return
	}
	g.cells[y][x] = cell{glyph: glyph, color: color, level: level}
}

func (g *grid) render() {
	var sb strings.Builder
	sb.Grow((g.width+12)*g.height + 8)
	sb.WriteString(term.Home)
	for _, row := range g.cells {
		lastColor := ""
		for _, c := range row {
			if c.color != lastColor {
				if c.color == "" {
					sb.WriteString(term.Reset)
				} else {
					sb.WriteString(c.color)
				}
				lastColor = c.color
			}
			sb.WriteByte(c.glyph)
		}
		sb.WriteString(term.Reset)
		sb.WriteByte('\n')
	}
	fmt.Print(sb.String())
}

type bubble struct {
	x, y   float64
	speed  float64
	wobble float64
	phase  float64
	size   int
}

type jelly struct {
	anchorX, anchorY float64
	halfWidth        int
	capHeight        int
	tentacleLength   int
	driftX           float64
	driftSpeed       float64
	phase            float64
	level            int
	palette          []string
}

type scene struct {
	width, height int
	jellies       []jelly
	bubbles       []bubble
	rng           *rand.Rand
}

func newScene(width, height int, rng *rand.Rand) *scene {
	primaryCap := min(8, max(5, height/5))
	primaryWidth := min(21, max(12, width/5))
	primaryTentacles := max(8, height-int(float64(height)*0.37)-3)

	jellies := []jelly{
		{
			anchorX:        0.50,
			anchorY:        0.37,
			halfWidth:      primaryWidth,
			capHeight:      primaryCap,
			tentacleLength: primaryTentacles,
			driftX:         0.14,
			driftSpeed:     0.017,
			phase:          0.3,
			level:          30,
			palette:        cyanBell,
		},
		{
			anchorX:        0.79,
			anchorY:        0.27,
			halfWidth:      max(7, primaryWidth/2),
			capHeight:      max(4, primaryCap/2),
			tentacleLength: max(8, primaryTentacles/2),
			driftX:         0.08,
			driftSpeed:     0.012,
			phase:          2.4,
			level:          20,
			palette:        pinkBell,
		},
	}
	if width >= 90 && height >= 30 {
		jellies = append(jellies, jelly{
			anchorX:        0.20,
			anchorY:        0.25,
			halfWidth:      max(6, primaryWidth/3),
			capHeight:      max(3, primaryCap/2),
			tentacleLength: max(7, primaryTentacles/2),
			driftX:         0.06,
			driftSpeed:     0.010,
			phase:          4.6,
			level:          10,
			palette:        blueBell,
		})
	}

	s := &scene{width: width, height: height, jellies: jellies, rng: rng}
	for i := 0; i < max(10, width*height/180); i++ {
		s.bubbles = append(s.bubbles, s.randomBubble(rng.Float64()*float64(height)))
	}
	return s
}

func (s *scene) randomBubble(y float64) bubble {
	return bubble{
		x:      s.rng.Float64() * float64(s.width),
		y:      y,
		speed:  0.08 + s.rng.Float64()*0.18,
		wobble: 0.3 + s.rng.Float64()*0.8,
		phase:  s.rng.Float64() * math.Pi * 2,
		size:   s.rng.Intn(4),
	}
}

// Run starts the jellyfish animation and continues until the process is interrupted.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := newScene(cfg.Width, cfg.Height, rng)
	g := newGrid(cfg.Width, cfg.Height)

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		g.clear()
		drawWater(g, frame)
		s.updateBubbles(frame)
		s.drawBubbles(g, frame)
		for i := len(s.jellies) - 1; i >= 0; i-- {
			drawJelly(g, s.jellies[i], frame)
		}
		g.render()
		<-ticker.C
	}
}

func drawWater(g *grid, frame int) {
	for y := 0; y < g.height; y++ {
		color := waterPalette[min(len(waterPalette)-1, y*len(waterPalette)/max(1, g.height))]
		for x := 0; x < g.width; x++ {
			// Sparse, deterministic flecks keep the water alive without flickering noise.
			v := math.Sin(float64(x)*0.19+float64(frame)*0.012) +
				math.Sin(float64(y)*0.47-float64(frame)*0.018)
			if (x*17+y*31)%97 == 0 && v > 0.45 {
				g.set(x, y, '.', color, 0)
			} else if (x*29+y*11+7)%211 == 0 && v < -0.65 {
				g.set(x, y, '`', color, 0)
			}
		}
	}
}

func (s *scene) updateBubbles(frame int) {
	for i := range s.bubbles {
		b := &s.bubbles[i]
		b.y -= b.speed
		b.x += math.Sin(float64(frame)*0.025+b.phase) * 0.025 * b.wobble
		if b.y < -1 || b.x < -1 || b.x > float64(s.width) {
			*b = s.randomBubble(float64(s.height) + s.rng.Float64()*4)
		}
	}
}

func (s *scene) drawBubbles(g *grid, frame int) {
	glyphs := [...]byte{'.', 'o', 'o', 'O'}
	for i, b := range s.bubbles {
		x := int(math.Round(b.x + math.Sin(float64(frame)*0.035+b.phase)*b.wobble))
		y := int(math.Round(b.y))
		color := bubblePalette[(i+b.size)%len(bubblePalette)]
		g.set(x, y, glyphs[b.size], color, 2)
	}
}

func drawJelly(g *grid, j jelly, frame int) {
	t := float64(frame)
	contraction := 1 + 0.055*math.Sin(t*0.105+j.phase)
	halfWidth := max(4, int(math.Round(float64(j.halfWidth)*contraction)))
	capHeight := max(3, int(math.Round(float64(j.capHeight)*(2-contraction))))
	cx := int(math.Round(j.anchorX*float64(g.width) +
		math.Sin(t*j.driftSpeed+j.phase)*j.driftX*float64(g.width)))
	cy := int(math.Round(j.anchorY*float64(g.height) +
		math.Sin(t*0.047+j.phase)*1.5))

	drawTentacles(g, j, cx, cy, halfWidth, frame)
	drawBellInterior(g, j, cx, cy, halfWidth, capHeight, frame)
	drawBellOutline(g, j, cx, cy, halfWidth, capHeight)
	drawSkirt(g, j, cx, cy, halfWidth, frame)
}

func drawBellInterior(g *grid, j jelly, cx, cy, halfWidth, capHeight, frame int) {
	for dy := -capHeight + 1; dy <= 0; dy++ {
		norm := float64(dy) / float64(capHeight)
		extent := int(float64(halfWidth-1) * math.Sqrt(max(0, 1-norm*norm)))
		for dx := -extent; dx <= extent; dx++ {
			if (dx*13+dy*23+frame/5+int(j.phase*10))%17 != 0 {
				continue
			}
			glyph := byte('.')
			if (dx*7-dy+frame/8)%5 == 0 {
				glyph = ':'
			}
			shade := min(len(j.palette)-2, max(0, (dy+capHeight)*3/max(1, capHeight)))
			g.set(cx+dx, cy+dy, glyph, j.palette[shade], j.level+2)
		}
	}
	// Radial ribs hint at the translucent structure inside the bell.
	for _, offset := range []int{-halfWidth / 2, 0, halfWidth / 2} {
		for step := 1; step < capHeight; step++ {
			y := cy - step
			ratio := float64(step) / float64(capHeight)
			x := cx + int(math.Round(float64(offset)*(1-ratio*ratio)))
			g.set(x, y, '|', j.palette[1], j.level+1)
		}
	}
}

func drawBellOutline(g *grid, j jelly, cx, cy, halfWidth, capHeight int) {
	steps := max(halfWidth*5, 32)
	for i := 0; i <= steps; i++ {
		theta := math.Pi + math.Pi*float64(i)/float64(steps)
		x := cx + int(math.Round(float64(halfWidth)*math.Cos(theta)))
		y := cy + int(math.Round(float64(capHeight)*math.Sin(theta)))
		cosine := math.Cos(theta)
		glyph := byte('_')
		if cosine < -0.28 {
			glyph = '/'
		} else if cosine > 0.28 {
			glyph = '\\'
		}
		colorIndex := 2 + int(math.Round(-math.Sin(theta)*float64(len(j.palette)-3)))
		g.set(x, y, glyph, j.palette[min(len(j.palette)-1, colorIndex)], j.level+8)
	}
}

func drawSkirt(g *grid, j jelly, cx, cy, halfWidth, frame int) {
	previousY := cy
	for dx := -halfWidth; dx <= halfWidth; dx++ {
		wave := math.Sin(float64(dx)*math.Pi*3/float64(max(1, halfWidth)) +
			float64(frame)*0.08 + j.phase)
		y := cy + 1 + int(math.Round(wave))
		glyph := byte('_')
		if y > previousY {
			glyph = '\\'
		} else if y < previousY {
			glyph = '/'
		}
		g.set(cx+dx, y, glyph, j.palette[3], j.level+9)
		if dx%(max(2, halfWidth/4)) == 0 && y+1 < g.height {
			g.set(cx+dx, y+1, 'v', j.palette[2], j.level+7)
		}
		previousY = y
	}
}

func drawTentacles(g *grid, j jelly, cx, cy, halfWidth, frame int) {
	count := min(9, max(5, halfWidth/3+2))
	for i := 0; i < count; i++ {
		fraction := float64(i+1)/float64(count+1)*1.55 - 0.775
		startX := cx + int(math.Round(fraction*float64(halfWidth)))
		length := j.tentacleLength - absInt(i-(count-1)/2)%3
		previousX := startX
		for step := 1; step <= length; step++ {
			progress := float64(step) / float64(max(1, length))
			lag := float64(step)*0.31 + float64(i)*0.72
			sway := math.Sin(float64(frame)*0.072-lag+j.phase) * (0.35 + progress*2.1)
			sway += math.Sin(float64(frame)*0.031-float64(step)*0.13+j.phase*1.7) * progress
			x := startX + int(math.Round(sway))
			y := cy + 1 + step
			glyph := byte('|')
			if x > previousX {
				glyph = '\\'
			} else if x < previousX {
				glyph = '/'
			}
			if step == length {
				glyph = ','
			}
			colorIndex := min(3, 1+int(progress*3))
			g.set(x, y, glyph, j.palette[colorIndex], j.level+4)
			previousX = x
		}
	}

	// A few wider oral arms make the center feel substantial, like the reference.
	for arm := -1; arm <= 1; arm++ {
		startX := cx + arm*max(2, halfWidth/4)
		length := max(5, j.tentacleLength*2/3)
		for step := 1; step <= length; step++ {
			progress := float64(step) / float64(length)
			x := startX + int(math.Round(math.Sin(float64(frame)*0.058-float64(step)*0.24+
				float64(arm)*1.4+j.phase)*(0.5+progress*1.7)))
			y := cy + 2 + step
			glyph := byte('}')
			if arm < 0 {
				glyph = '{'
			} else if arm == 0 {
				glyph = '#'
			}
			g.set(x, y, glyph, j.palette[2], j.level+5)
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
