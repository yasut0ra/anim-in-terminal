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
		"\x1b[38;5;18m",
		"\x1b[38;5;19m",
		"\x1b[38;5;20m",
		"\x1b[38;5;26m",
		"\x1b[38;5;27m",
		"\x1b[38;5;33m",
	}
	horizonPalette = []string{
		"\x1b[38;5;54m",
		"\x1b[38;5;55m",
		"\x1b[38;5;90m",
		"\x1b[38;5;129m",
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
	planktonPalette = []string{
		"\x1b[38;5;45m",
		"\x1b[38;5;81m",
		"\x1b[38;5;117m",
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
	vx    float64
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
	plankton := make([]bubble, 0, 128)

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		clearGrid(grid)
		drawSky(grid, frame)
		drawHorizonGlow(grid, frame)
		drawWaveLayers(grid, frame)
		drawFoam(grid, frame)
		updatePlankton(&plankton, cfg.Width, cfg.Height)
		drawPlankton(grid, plankton)
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
		idx := (y/2 + frame/18) % len(skyPalette)
		color := skyPalette[idx]
		for x := 0; x < width; x++ {
			grid[y][x] = cell{glyph: ' ', color: color}
		}
	}
	drawClouds(grid, frame)
}

func drawClouds(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	limit := height / 3
	for i := 0; i < width/6; i++ {
		x := (i*9 + frame/2) % width
		y := limit/2 + int(math.Sin(float64(x)/10+float64(frame)*0.01)*3)
		if y < 1 || y >= limit {
			continue
		}
		color := skyPalette[(i+frame/12)%len(skyPalette)]
		setIfEmpty(grid, x, y, '~', color)
		setIfEmpty(grid, (x+1)%width, y, '~', color)
	}
}

func drawHorizonGlow(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	line := height / 3
	for y := line; y < line+3 && y < height; y++ {
		color := horizonPalette[(y+frame/10)%len(horizonPalette)]
		for x := 0; x < width; x++ {
			setIfEmpty(grid, x, y, ' ', color)
		}
	}
}

func drawWaveLayers(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height / 3
	layerConfigs := []struct {
		scale float64
		speed float64
		amp   float64
	}{
		{scale: 1.0, speed: 1.0, amp: 1},
		{scale: 1.5, speed: 0.7, amp: 0.8},
		{scale: 2.3, speed: 0.4, amp: 0.6},
	}
	for y := base; y < height; y++ {
		py := float64(y-base) / float64(height-base)
		color := wavePalette[(int(py*float64(len(wavePalette)))+frame/15)%len(wavePalette)]
		for x := 0; x < width; x++ {
			fx := float64(x) / float64(width)
			value := 0.0
			for _, cfg := range layerConfigs {
				value += cfg.amp * waveValue(fx*cfg.scale, py*cfg.scale, frame, cfg.speed)
			}
			value = value / float64(len(layerConfigs))
			glyph := waveGlyph(value)
			grid[y][x] = cell{glyph: glyph, color: color}
		}
	}
}

func waveValue(fx, fy float64, frame int, speed float64) float64 {
	t := float64(frame) * 0.035 * speed
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
			vx:    rand.Float64()*0.2 - 0.1,
			vy:    -0.3 - rand.Float64()*0.4,
			life:  40 + rand.Intn(40),
			color: foamPalette[rand.Intn(len(foamPalette))],
		})
	}
	items := *bubbles
	dst := items[:0]
	for i := range items {
		items[i].x += items[i].vx
		items[i].y += items[i].vy
		items[i].life--
		if items[i].y < float64(height/3) || items[i].life <= 0 {
			continue
		}
		dst = append(dst, items[i])
	}
	*bubbles = dst
}

func drawPlankton(grid [][]cell, plankton []bubble) {
	for _, p := range plankton {
		x := int(math.Round(p.x))
		y := int(math.Round(p.y))
		if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
			continue
		}
		setCell(grid, x, y, '.', p.color)
	}
}

func updatePlankton(plankton *[]bubble, width, height int) {
	if rand.Intn(4) == 0 {
		*plankton = append(*plankton, bubble{
			x:     rand.Float64() * float64(width),
			y:     float64(height/2 + rand.Intn(height/2)),
			vx:    rand.Float64()*0.3 - 0.15,
			vy:    -rand.Float64() * 0.1,
			life:  80 + rand.Intn(80),
			color: planktonPalette[rand.Intn(len(planktonPalette))],
		})
	}
	items := *plankton
	dst := items[:0]
	for i := range items {
		items[i].x += items[i].vx
		items[i].y += items[i].vy
		items[i].life--
		if items[i].y < float64(height/3) || items[i].life <= 0 {
			continue
		}
		dst = append(dst, items[i])
	}
	*plankton = dst
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
