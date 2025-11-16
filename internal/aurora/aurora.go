package aurora

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
		"\x1b[38;5;17m",
		"\x1b[38;5;18m",
		"\x1b[38;5;19m",
		"\x1b[38;5;54m",
		"\x1b[38;5;55m",
	}
	auroraPalette = []string{
		"\x1b[38;5;35m",
		"\x1b[38;5;41m",
		"\x1b[38;5;47m",
		"\x1b[38;5;83m",
		"\x1b[38;5;119m",
		"\x1b[38;5;159m",
	}
	starPalette = []string{
		"\x1b[38;5;231m",
		"\x1b[38;5;195m",
		"\x1b[38;5;153m",
	}
	mountainPalette = []string{
		"\x1b[38;5;235m",
		"\x1b[38;5;236m",
		"\x1b[38;5;237m",
	}
)

// Config controls the aurora animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a typical terminal preset.
func DefaultConfig() Config {
	return Config{
		Width:      100,
		Height:     34,
		FrameDelay: 40 * time.Millisecond,
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
		c.FrameDelay = 45 * time.Millisecond
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

// Run launches the aurora animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	grid := newGrid(cfg.Width, cfg.Height)

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		clearGrid(grid)
		drawSky(grid, frame)
		drawStars(grid, frame)
		drawAuroraCurtains(grid, frame)
		drawMountains(grid, frame)
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
	for y := 0; y < height/2; y++ {
		color := skyPalette[(y/2+frame/30)%len(skyPalette)]
		for x := 0; x < width; x++ {
			grid[y][x] = cell{glyph: ' ', color: color}
		}
	}
}

func drawStars(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for i := 0; i < width/4; i++ {
		x := (i*17 + frame) % width
		y := rand.Intn(height / 2)
		color := starPalette[(x+y+frame/5)%len(starPalette)]
		if (x+y+frame)%13 == 0 {
			setCell(grid, x, y, '*', color)
		} else if (x*3+y+frame)%19 == 0 {
			setCell(grid, x, y, '+', color)
		}
	}
}

func drawAuroraCurtains(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height / 3
	for band := 0; band < 3; band++ {
		for x := 0; x < width; x++ {
			fx := float64(x) / float64(width)
			phase := float64(frame)*0.02 + float64(band)*1.1
			offset := math.Sin(fx*5+phase) * float64(6-band*2)
			y := base + band*3 + int(offset)
			if y < 0 || y >= height {
				continue
			}
			value := (math.Sin(fx*12+phase*1.5) + 1) / 2
			color := auroraPalette[(int(value*float64(len(auroraPalette)))+band)%len(auroraPalette)]
			glyph := curtainGlyph(value)
			setCell(grid, x, y, glyph, color)
			if y+1 < height && rand.Intn(3) == 0 {
				setCell(grid, x, y+1, glyph, color)
			}
		}
	}
}

func curtainGlyph(v float64) byte {
	switch {
	case v < 0.2:
		return '.'
	case v < 0.5:
		return '|'
	case v < 0.7:
		return '/'
	default:
		return '\\'
	}
}

func drawMountains(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height - 6
	for x := 0; x < width; x++ {
		offset := int(math.Sin(float64(x)/7+float64(frame)*0.005) * 4)
		y := base - offset
		color := mountainPalette[(x/5)%len(mountainPalette)]
		for dy := 0; dy < height-y; dy++ {
			if y+dy >= height {
				break
			}
			setIfEmpty(grid, x, y+dy, '#', color)
		}
	}
}

func setCell(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[0]) {
		return
	}
	grid[y][x] = cell{glyph: glyph, color: color}
}

func setIfEmpty(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) {
		return
	}
	if x < 0 || x >= len(grid[0]) {
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
