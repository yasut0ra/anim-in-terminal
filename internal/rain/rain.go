package rain

import (
	"fmt"
	"math"
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
	baseX      int
	head       float64
	speed      float64
	length     int
	paletteIdx int
	layer      int
	swayPhase  float64
}

type splash struct {
	x, y   float64
	vx, vy float64
	life   int
	color  string
}

// Run launches the rain animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	streams := makeStreams(cfg)
	splashes := make([]splash, 0, 128)
	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		grid := newGrid(cfg.Width, cfg.Height)
		drawMist(grid, frame)
		drawStreams(grid, streams, frame, &splashes)
		drawSplashes(grid, splashes)
		render(grid)
		updateSplashes(&splashes, cfg.Width, cfg.Height)
		updateStreams(streams, cfg.Width, cfg.Height)

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

func drawStreams(grid [][]cell, streams []stream, frame int, splashes *[]splash) {
	height := len(grid)
	width := len(grid[0])
	for _, s := range streams {
		palette := streamPalettes[s.paletteIdx%len(streamPalettes)]
		head := int(s.head)
		column := streamColumn(s, frame, width)
		for i := 0; i < s.length; i++ {
			y := head - i
			if y < 0 || y >= height {
				continue
			}
			var color string
			if i == 0 {
				color = glowPalette[(frame+y)%len(glowPalette)]
			} else {
				color = palette[min(i/2+s.layer, len(palette)-1)]
			}
			glyph := glyphPool[(frame+y+i)%len(glyphPool)]
			setCell(grid, column, y, glyph, color)
			if i == 0 && y >= height-2 {
				emitSplash(splashes, column, height)
			}
		}
	}
}

func streamColumn(s stream, frame int, width int) int {
	sway := math.Sin(s.swayPhase + float64(frame)*0.02*float64(s.layer+1))
	offset := int(math.Round(sway * float64(s.layer+1)))
	col := s.baseX + offset
	if col < 0 {
		return 0
	}
	if col >= width {
		return width - 1
	}
	return col
}

func emitSplash(splashes *[]splash, x int, height int) {
	count := 2 + rand.Intn(3)
	baseY := float64(height - 2)
	for i := 0; i < count; i++ {
		*splashes = append(*splashes, splash{
			x:     float64(x) + rand.Float64()*0.6 - 0.3,
			y:     baseY,
			vx:    rand.Float64()*0.8 - 0.4,
			vy:    -0.6 - rand.Float64()*0.7,
			life:  10 + rand.Intn(10),
			color: glowPalette[rand.Intn(len(glowPalette))],
		})
	}
}

func drawSplashes(grid [][]cell, splashes []splash) {
	for _, sp := range splashes {
		x := int(math.Round(sp.x))
		y := int(math.Round(sp.y))
		if y < 0 || y >= len(grid) {
			continue
		}
		if x < 0 || x >= len(grid[y]) {
			continue
		}
		setCell(grid, x, y, '\'', sp.color)
	}
}

func updateSplashes(splashes *[]splash, width, height int) {
	items := *splashes
	dst := items[:0]
	for i := range items {
		items[i].x += items[i].vx
		items[i].y += items[i].vy
		items[i].vy += 0.08
		items[i].life--
		if items[i].x < 0 || items[i].x >= float64(width) {
			continue
		}
		if items[i].y >= float64(height-1) {
			continue
		}
		if items[i].life <= 0 {
			continue
		}
		dst = append(dst, items[i])
	}
	*splashes = dst
}

func updateStreams(streams []stream, width, height int) {
	for i := range streams {
		streams[i].head += streams[i].speed
		if int(streams[i].head)-streams[i].length > height {
			resetStream(&streams[i], width, height, false)
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
		resetStream(&streams[i], cfg.Width, cfg.Height, true)
	}
	return streams
}

func resetStream(s *stream, width, height int, visible bool) {
	s.baseX = rand.Intn(width)
	s.length = clampInt(6+rand.Intn(height/2), 6, height)
	s.layer = rand.Intn(3)
	baseSpeed := 0.35 + float64(s.layer)*0.25
	s.speed = baseSpeed + rand.Float64()*0.6
	s.paletteIdx = rand.Intn(len(streamPalettes))
	s.swayPhase = rand.Float64() * math.Pi * 2
	if visible {
		s.head = rand.Float64() * float64(height)
	} else {
		s.head = -float64(rand.Intn(height))
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
