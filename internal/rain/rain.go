package rain

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	minWidth  = 48
	minHeight = 24
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	streamPalettes = [][]string{
		{"\x1b[38;5;159m", "\x1b[38;5;81m", "\x1b[38;5;42m", "\x1b[38;5;35m"},
		{"\x1b[38;5;120m", "\x1b[38;5;47m", "\x1b[38;5;40m", "\x1b[38;5;34m"},
		{"\x1b[38;5;123m", "\x1b[38;5;75m", "\x1b[38;5;43m", "\x1b[38;5;29m"},
	}
	glowPalette = []string{
		"\x1b[38;5;195m",
		"\x1b[38;5;229m",
	}
	mistPalette = []string{
		"\x1b[38;5;236m",
		"\x1b[38;5;237m",
	}
	glyphPool = []byte{'0', '1', '|', '/', '\\', '[', ']'}
)

// Config controls the rain animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
	Density    float64
}

// DefaultConfig returns a preset tuned for most terminals.
func DefaultConfig() Config {
	return Config{
		Width:      96,
		Height:     34,
		FrameDelay: 55 * time.Millisecond,
		Density:    0.18,
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
	if c.Density <= 0 {
		c.Density = 0.15
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

type stream struct {
	x          int
	head       float64
	speed      float64
	length     int
	paletteIdx int
}

// Run launches the rain animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	streams := makeStreams(cfg)
	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		grid := newGrid(cfg.Width, cfg.Height)
		drawMist(grid, frame)
		drawStreams(grid, streams, frame)
		render(grid)
		updateStreams(streams, cfg.Height)

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

func drawMist(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y++ {
		if (y+frame/3)%3 != 0 {
			continue
		}
		color := mistPalette[(y/2+frame/10)%len(mistPalette)]
		for x := (y + frame) % 6; x < width; x += 6 {
			setIfEmpty(grid, x, y, '.', color)
		}
	}
}

func drawStreams(grid [][]cell, streams []stream, frame int) {
	height := len(grid)
	for _, s := range streams {
		palette := streamPalettes[s.paletteIdx%len(streamPalettes)]
		head := int(s.head)
		for i := 0; i < s.length; i++ {
			y := head - i
			if y < 0 || y >= height {
				continue
			}
			var color string
			if i == 0 {
				color = glowPalette[(frame+y)%len(glowPalette)]
			} else {
				color = palette[min(i/2, len(palette)-1)]
			}
			glyph := glyphPool[(frame+y+i)%len(glyphPool)]
			setCell(grid, s.x, y, glyph, color)
		}
	}
}

func updateStreams(streams []stream, height int) {
	for i := range streams {
		streams[i].head += streams[i].speed
		if int(streams[i].head)-streams[i].length > height {
			resetStream(&streams[i], height)
			streams[i].head = -float64(rand.Intn(height))
		}
	}
}

func makeStreams(cfg Config) []stream {
	count := int(float64(cfg.Width) * cfg.Density)
	if count < 4 {
		count = 4
	}
	streams := make([]stream, count)
	for i := range streams {
		streams[i].x = rand.Intn(cfg.Width)
		resetStream(&streams[i], cfg.Height)
		streams[i].head = rand.Float64() * float64(cfg.Height)
	}
	return streams
}

func resetStream(s *stream, height int) {
	s.length = clampInt(6+rand.Intn(height/2), 6, height)
	s.speed = 0.4 + rand.Float64()*0.9
	s.paletteIdx = rand.Intn(len(streamPalettes))
	s.head = -float64(rand.Intn(height))
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

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
