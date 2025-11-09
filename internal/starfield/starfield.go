package starfield

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	minWidth       = 48
	minHeight      = 24
	minDepth       = 0.12
	backdropStride = 4
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	starPalette = []string{
		"\x1b[38;5;250m",
		"\x1b[38;5;252m",
		"\x1b[38;5;255m",
	}
	trailPalette = []string{
		"\x1b[38;5;240m",
		"\x1b[38;5;245m",
	}
	backdropPalette = []string{
		"\x1b[38;5;236m",
		"\x1b[38;5;235m",
	}
	glyphPalette = []byte{'.', '+', '*'}
)

// Config controls the starfield animation characteristics.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
	Density    float64
	WarpSpeed  float64
}

// DefaultConfig returns a sensible preset for most terminals.
func DefaultConfig() Config {
	return Config{
		Width:      96,
		Height:     32,
		FrameDelay: 40 * time.Millisecond,
		Density:    0.03,
		WarpSpeed:  0.012,
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
		c.FrameDelay = 45 * time.Millisecond
	}
	if c.Density <= 0 {
		c.Density = 0.02
	}
	if c.WarpSpeed <= 0 {
		c.WarpSpeed = 0.01
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

type star struct {
	x, y, z  float64
	velocity float64
	prevX    int
	prevY    int
	hasPrev  bool
}

// Run launches the starfield warp animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	stars := makeStars(cfg)
	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		grid := newGrid(cfg.Width, cfg.Height)
		drawBackdrop(grid, frame)
		drawStars(grid, stars, cfg, frame)
		render(grid)

		<-ticker.C
	}
}

func makeStars(cfg Config) []star {
	count := int(float64(cfg.Width*cfg.Height) * cfg.Density)
	if count < 32 {
		count = 32
	}
	stars := make([]star, count)
	for i := range stars {
		resetStar(&stars[i], cfg)
		stars[i].z = rand.Float64()*0.7 + 0.4
	}
	return stars
}

func resetStar(s *star, cfg Config) {
	s.x = rand.Float64()*2 - 1
	s.y = rand.Float64()*2 - 1
	s.z = rand.Float64()*1.1 + 0.5
	s.velocity = cfg.WarpSpeed * (0.7 + rand.Float64()*0.8)
	s.hasPrev = false
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

func drawBackdrop(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y += backdropStride {
		color := backdropPalette[(y/backdropStride+frame/20)%len(backdropPalette)]
		for x := (y/2 + frame) % 6; x < width; x += 6 {
			setIfEmpty(grid, x, y, '.', color)
		}
	}
	centerX := width / 2
	centerY := height / 2
	setIfEmpty(grid, centerX, centerY, '+', "\x1b[38;5;238m")
}

func drawStars(grid [][]cell, stars []star, cfg Config, frame int) {
	width := len(grid[0])
	height := len(grid)
	for i := range stars {
		px, py, ok := projectStar(stars[i], width, height)
		if !ok {
			resetStar(&stars[i], cfg)
			continue
		}

		if stars[i].hasPrev {
			drawTrail(grid, stars[i].prevX, stars[i].prevY, px, py, stars[i].z)
		}

		color := starColor(stars[i].z, frame)
		glyph := starGlyph(stars[i].z)
		setCell(grid, px, py, glyph, color)

		stars[i].prevX = px
		stars[i].prevY = py
		stars[i].hasPrev = true

		stars[i].z -= stars[i].velocity
		if stars[i].z <= minDepth {
			resetStar(&stars[i], cfg)
		}
	}
}

func projectStar(s star, width, height int) (int, int, bool) {
	scale := float64(min(width, height)) * 0.45
	if s.z <= 0 {
		return 0, 0, false
	}
	x := int(float64(width)/2 + s.x*scale/s.z)
	y := int(float64(height)/2 + s.y*scale/(s.z*0.9))
	if x < 0 || x >= width || y < 0 || y >= height {
		return 0, 0, false
	}
	return x, y, true
}

func drawTrail(grid [][]cell, x0, y0, x1, y1 int, depth float64) {
	points := linePoints(x0, y0, x1, y1)
	if len(points) <= 1 {
		return
	}
	colorIndex := clampInt(int((1-depth)*float64(len(trailPalette))), 0, len(trailPalette)-1)
	color := trailPalette[colorIndex]
	glyph := drawTrailChar(depth)
	for idx := 0; idx < len(points)-1; idx++ {
		p := points[idx]
		setIfEmpty(grid, p[0], p[1], glyph, color)
	}
}

func starColor(depth float64, frame int) string {
	if len(starPalette) == 0 {
		return ""
	}
	ratio := clampFloat(1-depth, 0, 0.95)
	index := int(ratio / 0.4)
	if index >= len(starPalette) {
		index = len(starPalette) - 1
	}
	offset := (frame / 24) % len(starPalette)
	return starPalette[(index+offset)%len(starPalette)]
}

func starGlyph(depth float64) byte {
	if len(glyphPalette) == 0 {
		return '*'
	}
	ratio := clampFloat(1-depth, 0, 1)
	index := int(ratio * float64(len(glyphPalette)))
	if index >= len(glyphPalette) {
		index = len(glyphPalette) - 1
	}
	return glyphPalette[index]
}

func drawTrailChar(depth float64) byte {
	if depth > 0.6 {
		return '.'
	}
	if depth > 0.3 {
		return '-'
	}
	return '~'
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

func linePoints(x0, y0, x1, y1 int) [][2]int {
	points := make([][2]int, 0, max(abs(x1-x0), abs(y1-y0))+1)
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy

	for {
		points = append(points, [2]int{x0, y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
	return points
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
