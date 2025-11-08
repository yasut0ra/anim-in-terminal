package cloud

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	minWidthCloud  = 60
	minHeightCloud = 24
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	skyPalette = []string{
		"\x1b[38;5;111m",
		"\x1b[38;5;75m",
		"\x1b[38;5;45m",
		"\x1b[38;5;39m",
	}
	highCloudColors = []string{
		"\x1b[38;5;255m",
		"\x1b[38;5;252m",
	}
	midCloudColors = []string{
		"\x1b[38;5;250m",
		"\x1b[38;5;248m",
		"\x1b[38;5;246m",
	}
	lowCloudColors = []string{
		"\x1b[38;5;245m",
		"\x1b[38;5;243m",
		"\x1b[38;5;239m",
	}
	lightningPalette = []string{
		"\x1b[38;5;231m",
		"\x1b[38;5;229m",
		"\x1b[38;5;227m",
	}
)

// Config describes the cloud animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a preset suited for most terminals.
func DefaultConfig() Config {
	return Config{
		Width:      100,
		Height:     34,
		FrameDelay: 70 * time.Millisecond,
	}
}

func (c Config) normalize() Config {
	if c.Width < minWidthCloud {
		c.Width = minWidthCloud
	}
	if c.Height < minHeightCloud {
		c.Height = minHeightCloud
	}
	if c.FrameDelay <= 0 {
		c.FrameDelay = 70 * time.Millisecond
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

type cloudLayer struct {
	height    float64
	thickness float64
	density   float64
	scale     float64
	speed     float64
	colorSet  []string
	glyphs    []byte
	parallax  float64
}

type point struct {
	x int
	y int
}

type lightning struct {
	points []point
	life   int
}

// Run starts the cloud animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	layers := []cloudLayer{
		{
			height:    0.22,
			thickness: 0.18,
			density:   0.75,
			scale:     0.11,
			speed:     0.022,
			colorSet:  highCloudColors,
			glyphs:    []byte{'@', '%'},
			parallax:  0.7,
		},
		{
			height:    0.38,
			thickness: 0.22,
			density:   0.62,
			scale:     0.07,
			speed:     0.015,
			colorSet:  midCloudColors,
			glyphs:    []byte{'#', '*'},
			parallax:  0.9,
		},
		{
			height:    0.55,
			thickness: 0.28,
			density:   0.48,
			scale:     0.05,
			speed:     0.01,
			colorSet:  lowCloudColors,
			glyphs:    []byte{'=', '-'},
			parallax:  1.2,
		},
	}

	var bolt lightning

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		grid := newGrid(cfg.Width, cfg.Height)
		drawSky(grid)
		for i := range layers {
			drawLayer(grid, &layers[i], frame)
		}
		if !bolt.active() && rand.Float64() < 0.02 {
			bolt = newLightning(cfg.Width, cfg.Height)
		}
		if bolt.active() {
			drawLightning(grid, &bolt)
			bolt.life--
		}
		render(grid)
		<-ticker.C
	}
}

func newGrid(width, height int) [][]cell {
	grid := make([][]cell, height)
	for y := range grid {
		grid[y] = make([]cell, width)
		for x := range grid[y] {
			grid[y][x] = cell{glyph: ' ', color: ""}
		}
	}
	return grid
}

func drawSky(grid [][]cell) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y++ {
		color := skyPalette[min(len(skyPalette)-1, y*len(skyPalette)/max(1, height))]
		for x := 0; x < width; x++ {
			grid[y][x] = cell{glyph: '.', color: color}
		}
	}
}

func drawLayer(grid [][]cell, layer *cloudLayer, frame int) {
	height := len(grid)
	width := len(grid[0])
	if len(layer.glyphs) == 0 || len(layer.colorSet) == 0 {
		return
	}

	basePhase := float64(frame) * layer.speed
	for y := 0; y < height; y++ {
		yNorm := float64(y) / float64(height-1)
		distance := math.Abs(yNorm - layer.height)
		falloff := math.Exp(-math.Pow(distance/layer.thickness, 2) * 2.5)
		if falloff < 0.05 {
			continue
		}
		for x := 0; x < width; x++ {
			noise := cloudNoise(float64(x), float64(y), basePhase, layer)
			coverage := falloff*(0.55+0.45*noise) - (1-layer.density)*0.4
			if coverage < 0.35 {
				continue
			}
			glyph := layer.glyphs[0]
			if coverage < 0.55 && len(layer.glyphs) > 1 {
				glyph = layer.glyphs[1]
			}
			color := layer.colorSet[(x+y)%len(layer.colorSet)]
			setCell(grid, x, y, glyph, color)
		}
	}
}

func cloudNoise(x, y float64, phase float64, layer *cloudLayer) float64 {
	s := layer.scale
	p := layer.parallax
	v := math.Sin((x*s+p*phase)*0.9+phase*2.0) +
		0.6*math.Sin((x*0.35+y*0.25)*s*1.4-phase*1.2) +
		0.4*math.Sin((y*s*0.6-x*0.22)*0.8+phase*0.7)
	return math.Tanh(v)
}

func drawLightning(grid [][]cell, bolt *lightning) {
	for i, pt := range bolt.points {
		if pt.y < 0 || pt.y >= len(grid) || pt.x < 0 || pt.x >= len(grid[pt.y]) {
			continue
		}
		color := lightningPalette[min(len(lightningPalette)-1, i%len(lightningPalette))]
		setCell(grid, pt.x, pt.y, lightningGlyph(i), color)
	}
}

func lightningGlyph(i int) byte {
	switch i % 3 {
	case 0:
		return '|'
	case 1:
		return '/'
	default:
		return '\\'
	}
}

func newLightning(width, height int) lightning {
	points := make([]point, 0, height)
	x := rand.Intn(width/2) + width/4
	y := rand.Intn(height/6) + 1
	length := height/2 + rand.Intn(height/3)
	for i := 0; i < length && y < height-2; i++ {
		points = append(points, point{x: x, y: y})
		x += rand.Intn(3) - 1
		if x < 1 {
			x = 1
		}
		if x >= width-1 {
			x = width - 2
		}
		y += 1 + rand.Intn(2)
	}
	return lightning{points: points, life: 4 + rand.Intn(4)}
}

func (l lightning) active() bool {
	return l.life > 0 && len(l.points) > 0
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

func setCell(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = cell{glyph: glyph, color: color}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
