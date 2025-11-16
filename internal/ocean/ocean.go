package ocean

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	skyPalette = []string{
		"\x1b[38;5;24m",
		"\x1b[38;5;25m",
		"\x1b[38;5;26m",
		"\x1b[38;5;27m",
	}
	wavePalette = []string{
		"\x1b[38;5;30m",
		"\x1b[38;5;31m",
		"\x1b[38;5;37m",
		"\x1b[38;5;44m",
		"\x1b[38;5;51m",
	}
	foamPalette = []string{
		"\x1b[38;5;189m",
		"\x1b[38;5;195m",
		"\x1b[38;5;231m",
	}
)

// Config for ocean currents animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a preset that fits most terminals.
func DefaultConfig() Config {
	return Config{
		Width:      100,
		Height:     34,
		FrameDelay: 35 * time.Millisecond,
	}
}

func (c Config) normalize() Config {
	if c.Width < 60 {
		c.Width = 60
	}
	if c.Height < 24 {
		c.Height = 24
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

type bubble struct {
	x, y  float64
	vy    float64
	life  int
	color string
}

// Run starts the ocean currents animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	grid := newGrid(cfg.Width, cfg.Height)
	bubbles := make([]bubble, 0, 128)

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		clearGrid(grid)
		drawSky(grid, frame)
		drawWaves(grid, frame)
		drawFoam(grid, frame)
		updateBubbles(&bubbles, cfg.Width, cfg.Height)
		drawBubbles(grid, bubbles)
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

func clearGrid(grid [][]cell) {
	for y := range grid {
		for x := range grid[y] {
			grid[y][x] = cell{glyph: ' ', color: ""}
		}
	}
}

func drawSky(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	limit := height / 3
	for y := 0; y < limit; y++ {
		color := skyPalette[(y/2+frame/20)%len(skyPalette)]
		for x := 0; x < width; x++ {
			grid[y][x] = cell{glyph: ' ', color: color}
		}
	}
}

func drawWaves(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height / 3
	for y := base; y < height; y++ {
		py := float64(y-base) / float64(height-base)
		color := wavePalette[(int(py*float64(len(wavePalette)))+frame/15)%len(wavePalette)]
		for x := 0; x < width; x++ {
			fx := float64(x) / float64(width)
			value := waveValue(fx, py, frame)
			glyph := waveGlyph(value)
			grid[y][x] = cell{glyph: glyph, color: color}
		}
	}
}

func waveValue(fx, fy float64, frame int) float64 {
	t := float64(frame) * 0.035
	value := math.Sin((fx*8+fy*6)*math.Pi+t) +
		0.7*math.Sin((fx*3-fy*5)*math.Pi+t*0.7) +
		0.5*math.Sin((fx+fy)*12*math.Pi+t*1.4)
	return (value + 3) / 6
}

func waveGlyph(v float64) byte {
	switch {
	case v < 0.2:
		return '`'
	case v < 0.4:
		return '.'
	case v < 0.6:
		return '-'
	case v < 0.8:
		return '='
	default:
		return '~'
	}
}

func drawFoam(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height - 5
	for x := 0; x < width; x++ {
		if (x+frame)%7 == 0 {
			color := foamPalette[(x/4+frame/10)%len(foamPalette)]
			for dy := 0; dy < 2 && base-dy >= height/3; dy++ {
				setIfEmpty(grid, x, base-dy, '*', color)
			}
		}
	}
}

func drawBubbles(grid [][]cell, bubbles []bubble) {
	for _, b := range bubbles {
		x := int(math.Round(b.x))
		y := int(math.Round(b.y))
		if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
			continue
		}
		setCell(grid, x, y, 'o', b.color)
	}
}

func updateBubbles(bubbles *[]bubble, width, height int) {
	if rand.Intn(3) == 0 {
		*bubbles = append(*bubbles, bubble{
			x:     rand.Float64() * float64(width),
			y:     float64(height - 1),
			vy:    -0.3 - rand.Float64()*0.4,
			life:  40 + rand.Intn(40),
			color: foamPalette[rand.Intn(len(foamPalette))],
		})
	}
	items := *bubbles
	dst := items[:0]
	for i := range items {
		items[i].y += items[i].vy
		items[i].life--
		if items[i].y < float64(height/3) || items[i].life <= 0 {
			continue
		}
		dst = append(dst, items[i])
	}
	*bubbles = dst
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

func setIfEmpty(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[y]) {
		return
	}
	if grid[y][x].glyph == ' ' {
		grid[y][x] = cell{glyph: glyph, color: color}
	}
}

func render(grid [][]cell) {
	var sb strings.Builder
	height := len(grid)
	width := len(grid[0])
	sb.Grow((width+8)*height + 16)
	sb.WriteString(ansiHome)
	for _, row := range grid {
		for _, c := range row {
			if c.color != "" {
				sb.WriteString(c.color)
			}
			sb.WriteByte(c.glyph)
		}
		sb.WriteString(ansiReset)
		sb.WriteByte('\n')
	}
	fmt.Print(sb.String())
}
