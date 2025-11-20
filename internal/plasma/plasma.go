package plasma

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	minWidth     = 60
	minHeight    = 24
	paletteSize  = 12
	glowStrength = 0.18
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
		"\x1b[38;5;51m",
		"\x1b[38;5;87m",
		"\x1b[38;5;123m",
		"\x1b[38;5;159m",
		"\x1b[38;5;195m",
	}
	glyphPalette = []byte{' ', '.', ',', ':', '-', '=', '*', '#', '%', '@'}
)

// Config controls the plasma animation behaviour.
type Config struct {
	Width         int
	Height        int
	FrameDelay    time.Duration
	PaletteScroll float64
}

// DefaultConfig returns sane defaults for typical terminals.
func DefaultConfig() Config {
	return Config{
		Width:         100,
		Height:        34,
		FrameDelay:    35 * time.Millisecond,
		PaletteScroll: 0.08,
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
	if c.PaletteScroll <= 0 {
		c.PaletteScroll = 0.05
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

// Run launches the plasma grid animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	grid := newGrid(cfg.Width, cfg.Height)

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		drawPlasma(grid, frame, cfg)
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

func drawPlasma(grid [][]cell, frame int, cfg Config) {
	height := len(grid)
	width := len(grid[0])
	t := float64(frame) * 0.03
	scroll := float64(frame) * cfg.PaletteScroll

	for y := 0; y < height; y++ {
		fy := float64(y) / float64(height)
		for x := 0; x < width; x++ {
			fx := float64(x) / float64(width)
			value := plasmaValue(fx, fy, t)
			color := paletteForValue(value + scroll)
			glyph := glyphForValue(value)
			grid[y][x] = cell{glyph: glyph, color: color}
		}
	}

	drawScanline(grid, frame)
	drawGlow(grid, frame)
}

func plasmaValue(fx, fy, t float64) float64 {
	v := math.Sin((fx*10)+t) +
		math.Sin((fy*12)-t*0.7) +
		math.Sin((fx+fy)*8+t*0.3) +
		0.5*math.Sin(math.Hypot(fx-0.5, fy-0.5)*15-t*1.5)

	noise := simpleNoise(fx, fy, t)
	return (v/3.5 + noise*0.25 + 1) / 2 // normalize 0..1
}

func simpleNoise(x, y, t float64) float64 {
	n := math.Sin((x*13+y*17+t)*12.9898) * 43758.5453
	return math.Mod(math.Abs(n), 1)
}

func paletteForValue(v float64) string {
	if len(colorPalette) == 0 {
		return ""
	}
	v = math.Mod(v, float64(len(colorPalette)))
	if v < 0 {
		v += float64(len(colorPalette))
	}
	idx := int(v) % len(colorPalette)
	return colorPalette[idx]
}

func glyphForValue(v float64) byte {
	if len(glyphPalette) == 0 {
		return '#'
	}
	idx := int(clampFloat(v*float64(len(glyphPalette)), 0, float64(len(glyphPalette)-1)))
	return glyphPalette[idx]
}

func drawScanline(grid [][]cell, frame int) {
	height := len(grid)
	if height == 0 {
		return
	}
	y := (frame / 3) % height
	for x := 0; x < len(grid[y]); x++ {
		grid[y][x].color = "\x1b[38;5;231m"
		if grid[y][x].glyph == ' ' {
			grid[y][x].glyph = '-'
		}
	}
}

func drawGlow(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	centerX := float64(width) / 2
	centerY := float64(height) / 2
	pulse := 0.5 + 0.5*math.Sin(float64(frame)*0.04)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dist := math.Hypot(float64(x)-centerX, (float64(y)-centerY)*0.8)
			falloff := math.Exp(-dist * glowStrength)
			if falloff < 0.1 {
				continue
			}
			boost := pulse * falloff
			color := paletteForValue(boost * float64(len(colorPalette)))
			grid[y][x].color = color
		}
	}
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

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
