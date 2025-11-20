package tunnel

import (
	"fmt"
	"math"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	minWidth  = 60
	minHeight = 24
)

var (
	colorPalette = []string{
		"\x1b[38;5;17m",
		"\x1b[38;5;18m",
		"\x1b[38;5;19m",
		"\x1b[38;5;20m",
		"\x1b[38;5;27m",
		"\x1b[38;5;33m",
		"\x1b[38;5;39m",
		"\x1b[38;5;45m",
		"\x1b[38;5;51m",
		"\x1b[38;5;87m",
		"\x1b[38;5;123m",
		"\x1b[38;5;159m",
		"\x1b[38;5;195m",
	}
	glyphPalette = []byte{' ', '.', '.', ':', '-', '+', '*', 'x', 'X', '#', '@'}
	starPalette  = []string{
		"\x1b[38;5;25m",
		"\x1b[38;5;31m",
		"\x1b[38;5;33m",
		"\x1b[38;5;39m",
		"\x1b[38;5;45m",
		"\x1b[38;5;51m",
	}
	accentPalette = []string{
		"\x1b[38;5;51m",
		"\x1b[38;5;87m",
		"\x1b[38;5;123m",
		"\x1b[38;5;159m",
	}
)

// Config controls the tunnel animation behaviour.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns sane defaults for typical terminals.
func DefaultConfig() Config {
	return Config{
		Width:      100,
		Height:     34,
		FrameDelay: 35 * time.Millisecond,
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
		c.FrameDelay = 40 * time.Millisecond
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

// Run launches the neon tunnel animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	grid := newGrid(cfg.Width, cfg.Height)

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		drawTunnel(grid, frame)
		render(grid)
		<-ticker.C
	}
}

func newGrid(width, height int) [][]cell {
	grid := make([][]cell, height)
	for y := range grid {
		grid[y] = make([]cell, width)
	}
	return grid
}

func drawTunnel(grid [][]cell, frame int) {
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])

	t := float64(frame) * 0.045
	swirl := float64(frame) * 0.02
	depthPulse := 0.55 + 0.4*math.Sin(float64(frame)*0.05)

	for y := 0; y < height; y++ {
		ny := (float64(y)/float64(height) - 0.5) * 2
		ny *= 0.72
		for x := 0; x < width; x++ {
			nx := (float64(x)/float64(width) - 0.5) * 2
			nx *= 1.1

			r := math.Hypot(nx, ny) + 0.0001
			angle := math.Atan2(ny, nx)

			depth := 1.0 / (r*2.2 + 0.5)
			wave := math.Sin(1.5/r - t*1.7 + math.Cos(angle*3+swirl)*0.55)
			spiral := math.Sin(angle*6 + t*2.1)
			flow := math.Cos(r*14 - t*3.4 + angle*1.3)
			band := math.Cos((r-depthPulse)*9 - t*1.2)

			value := wave*0.62 + spiral*0.24 + flow*0.28 + band*0.18 - r*0.95
			intensity := value + depth*0.9

			grid[y][x] = cell{
				glyph: glyphForValue(intensity),
				color: paletteForValue(intensity),
			}
		}
	}

	drawBackgroundStars(grid, frame)
	drawRays(grid, frame)
	drawDebris(grid, frame)
	drawPulseRings(grid, frame)
	drawCenterGlow(grid, frame)
}

func drawCenterGlow(grid [][]cell, frame int) {
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])
	cx := width / 2
	cy := height / 2

	radius := 1 + int(2*(0.5+0.5*math.Sin(float64(frame)*0.1+1.4)))
	for y := cy - radius; y <= cy+radius; y++ {
		if y < 0 || y >= height {
			continue
		}
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || x >= width {
				continue
			}
			dist := math.Hypot(float64(x-cx), float64(y-cy))
			if dist <= float64(radius) {
				grid[y][x] = cell{glyph: '*', color: "\x1b[38;5;195m"}
			}
		}
	}
}

func drawPulseRings(grid [][]cell, frame int) {
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])
	cx := width / 2
	cy := height / 2
	maxR := float64(width)/2 - 1
	if maxR < 2 {
		return
	}

	aspect := 1.0
	speed := 1.15
	thickness := 1.8
	gap := 10.0
	cycle := maxR + thickness*2 + gap
	phase := math.Mod(float64(frame)*speed, cycle)
	if phase > maxR+thickness {
		return
	}
	radius := math.Min(maxR, math.Max(1, phase))
	color := accentPalette[(frame/7)%len(accentPalette)]

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x - cx)
			dy := float64(y-cy) * aspect
			dist := math.Hypot(dx, dy)
			band := math.Abs(dist - radius)
			if band > thickness {
				continue
			}
			intensity := clamp(1-(band/thickness), 0, 1)
			glyph := '.'
			if intensity > 0.65 {
				glyph = '*'
			}
			grid[y][x] = cell{glyph: byte(glyph), color: color}
		}
	}
}

func drawBackgroundStars(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y += 2 {
		for x := (y + frame/3) % 6; x < width; x += 6 {
			color := starPalette[(x/3+y+frame/11)%len(starPalette)]
			if ((x*37 + y*13 + frame) % 57) < 3 {
				setCell(grid, x, y, '.', color)
			} else if ((x*19 + y*7 + frame*2) % 71) == 0 {
				setCell(grid, x, y, '+', color)
			}
		}
	}
}

func drawRays(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	cx := width / 2
	cy := height / 2
	count := 14
	maxR := float64(width) / 2
	for i := 0; i < count; i++ {
		angle := float64(i)/float64(count)*math.Pi*2 + math.Sin(float64(frame)*0.012)*0.6
		phase := math.Sin(float64(frame)*0.06+float64(i)) * 0.5
		length := maxR * (0.6 + 0.35*phase)
		color := accentPalette[(i+frame/6)%len(accentPalette)]
		for r := 1.0; r < length; r += 0.8 {
			x := cx + int(math.Cos(angle)*r)
			y := cy + int(math.Sin(angle)*r*0.6)
			if x < 0 || x >= width || y < 0 || y >= height {
				continue
			}
			glyph := '|'
			if i%2 == 0 {
				glyph = '/'
			}
			setCell(grid, x, y, byte(glyph), color)
		}
	}
}

func drawDebris(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	cx := width / 2
	cy := height / 2
	count := width / 2
	for i := 0; i < count; i++ {
		f := float64(i) + float64(frame)*0.9
		theta := math.Sin(f*0.03+float64(frame)*0.001)*math.Pi + float64(i%7)*0.4
		r := math.Mod(f*0.18, float64(width)/2) * (0.7 + 0.3*math.Sin(float64(frame)*0.02))
		x := cx + int(math.Cos(theta)*r)
		y := cy + int(math.Sin(theta)*r*0.65)
		if x < 0 || x >= width || y < 0 || y >= height {
			continue
		}
		color := colorPalette[(i+frame/5)%len(colorPalette)]
		glyph := glyphPalette[(i+frame)%len(glyphPalette)]
		setCell(grid, x, y, glyph, color)
	}
}

func paletteForValue(v float64) string {
	if len(colorPalette) == 0 {
		return ""
	}
	norm := clamp((v+1.3)/2.6, 0, 0.9999)
	idx := int(norm * float64(len(colorPalette)))
	return colorPalette[idx]
}

func glyphForValue(v float64) byte {
	if len(glyphPalette) == 0 {
		return '#'
	}
	norm := clamp((v+1.0)/2.0, 0, 0.9999)
	idx := int(norm * float64(len(glyphPalette)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(glyphPalette) {
		idx = len(glyphPalette) - 1
	}
	return glyphPalette[idx]
}

func setCell(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = cell{glyph: glyph, color: color}
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func render(grid [][]cell) {
	var sb strings.Builder
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])
	sb.Grow((width+8)*height + 16)
	sb.WriteString(term.Home)

	for _, row := range grid {
		for _, c := range row {
			if c.color != "" {
				sb.WriteString(c.color)
			}
			g := c.glyph
			if g == 0 {
				g = ' '
			}
			sb.WriteByte(g)
		}
		sb.WriteString(term.Reset)
		sb.WriteByte('\n')
	}

	fmt.Print(sb.String())
}
