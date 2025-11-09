package skyline

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
		"\x1b[38;5;20m",
		"\x1b[38;5;26m",
	}
	horizonPalette = []string{
		"\x1b[38;5;90m",
		"\x1b[38;5;129m",
		"\x1b[38;5;165m",
		"\x1b[38;5;201m",
	}
	buildingPalettes = [][]string{
		{"\x1b[38;5;236m", "\x1b[38;5;237m", "\x1b[38;5;238m"},
		{"\x1b[38;5;60m", "\x1b[38;5;61m", "\x1b[38;5;62m"},
		{"\x1b[38;5;33m", "\x1b[38;5;39m", "\x1b[38;5;45m"},
		{"\x1b[38;5;129m", "\x1b[38;5;135m", "\x1b[38;5;141m"},
	}
	windowPalette = []string{
		"\x1b[38;5;226m",
		"\x1b[38;5;190m",
		"\x1b[38;5;214m",
		"\x1b[38;5;51m",
	}
	glowPalette = []string{
		"\x1b[38;5;33m",
		"\x1b[38;5;45m",
		"\x1b[38;5;51m",
	}
)

// Config controls the skyline animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a preset that works for most terminals.
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
	color string
	glyph byte
}

type building struct {
	x         int
	width     int
	height    int
	palette   []string
	layer     int
	windowOn  []bool
	outline   string
	fillGlyph byte
}

// Run starts the neon skyline animation.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	grid := newGrid(cfg.Width, cfg.Height)
	buildings := makeBuildings(cfg)

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		clearGrid(grid)
		drawSky(grid, frame)
		drawStars(grid, frame)
		drawHorizonGlow(grid, frame)
		drawBuildings(grid, buildings, frame)
		drawHUD(grid, frame)
		render(grid)

		updateBuildings(buildings, cfg.Width, frame)

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
		hue := skyPalette[(y/2+frame/20)%len(skyPalette)]
		for x := 0; x < width; x++ {
			grid[y][x] = cell{glyph: ' ', color: hue}
		}
	}
}

func drawStars(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for i := 0; i < width/6; i++ {
		x := (i*13 + frame) % width
		y := (i*7 + frame/3) % (height / 2)
		if (x+y+frame)%11 == 0 {
			grid[y][x] = cell{glyph: '.', color: "\x1b[38;5;231m"}
		} else if (x*3+y+frame)%17 == 0 {
			grid[y][x] = cell{glyph: '+', color: "\x1b[38;5;81m"}
		}
	}
}

func drawHorizonGlow(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	horizon := height / 2
	for y := horizon; y < height; y++ {
		falloff := float64(y-horizon) / float64(height-horizon)
		color := horizonPalette[(int(falloff*float64(len(horizonPalette)))+frame/16)%len(horizonPalette)]
		for x := 0; x < width; x++ {
			if grid[y][x].glyph == ' ' {
				grid[y][x] = cell{glyph: ' ', color: color}
			}
		}
	}
}

func makeBuildings(cfg Config) []building {
	layers := []int{3, 2, 1}
	result := make([]building, 0, cfg.Width/2)
	for _, layer := range layers {
		x := rand.Intn(8)
		for x < cfg.Width {
			width := 4 + rand.Intn(6+layer*2)
			height := cfg.Height/4 + rand.Intn(cfg.Height/4) + layer*3
			palette := buildingPalettes[rand.Intn(len(buildingPalettes))]
			windowCount := width * height / 5
			windows := make([]bool, windowCount)
			for i := range windows {
				chance := max(1, 3-layer)
				windows[i] = rand.Intn(chance) == 0
			}
			fillGlyph := []byte{'=', '#', '%'}[min(layer, 3)-1]
			outline := glowPalette[rand.Intn(len(glowPalette))]
			result = append(result, building{
				x:         x,
				width:     width,
				height:    height,
				palette:   palette,
				layer:     layer,
				windowOn:  windows,
				outline:   outline,
				fillGlyph: fillGlyph,
			})
			x += width + rand.Intn(6)
		}
	}
	return result
}

func drawBuildings(grid [][]cell, buildings []building, frame int) {
	baseLine := len(grid) - 3
	for _, layer := range []int{3, 2, 1} {
		for _, b := range buildings {
			if b.layer == layer {
				drawBuilding(grid, b, baseLine, frame)
			}
		}
	}
}

func drawBuilding(grid [][]cell, b building, baseLine int, frame int) {
	height := b.height
	top := baseLine - height
	if top < 0 {
		top = 0
	}
	layerOffset := b.layer
	for y := 0; y < height && top+y < len(grid); y++ {
		color := b.palette[(y+layerOffset)%len(b.palette)]
		for x := 0; x < b.width; x++ {
			col := b.x + x
			if col < 0 || col >= len(grid[0]) {
				continue
			}
			var glyph byte = b.fillGlyph
			edgeColor := color
			if x == 0 || x == b.width-1 {
				glyph = '|'
				edgeColor = b.outline
			} else if y == 0 {
				glyph = '_'
				edgeColor = b.outline
			}
			grid[top+y][col] = cell{glyph: glyph, color: edgeColor}
		}
	}
	drawWindows(grid, b, baseLine, frame)
	drawBillboard(grid, b, baseLine, frame)
}

func drawWindows(grid [][]cell, b building, baseLine int, frame int) {
	windowCols := max(1, b.width/2)
	windowRows := max(2, b.height/4)
	idx := 0
	for wy := 0; wy < windowRows; wy++ {
		y := baseLine - wy*3 - 2
		if y <= 2 {
			continue
		}
		for wx := 0; wx < windowCols; wx++ {
			if idx >= len(b.windowOn) {
				break
			}
			if b.windowOn[idx] || (frame/10+wx+wy)%6 == 0 {
				x := b.x + 1 + wx*2
				color := windowPalette[(wx+wy+frame/7)%len(windowPalette)]
				setCell(grid, x, y, ':', color)
				setCell(grid, x+1, y, ':', color)
			}
			idx++
		}
	}
}

func drawBillboard(grid [][]cell, b building, baseLine int, frame int) {
	if b.width < 8 {
		return
	}
	y := baseLine - b.height - 3
	if y < 1 {
		return
	}
	x := b.x + b.width/2 - 4
	for i := 0; i < 8; i++ {
		color := glowPalette[(i+frame/6)%len(glowPalette)]
		setCell(grid, x+i, y, '-', color)
		setCell(grid, x+i, y+1, '-', color)
	}
	if (frame/40)%2 == 0 {
		color := "\x1b[38;5;219m"
		setCell(grid, x+2, y-1, '/', color)
		setCell(grid, x+5, y-1, '\\', color)
	}
}

func drawHUD(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	y := height - 2
	barWidth := width / 2
	start := (width - barWidth) / 2
	fill := int(float64(barWidth) * (0.5 + 0.5*math.Sin(float64(frame)*0.02)))
	for x := 0; x < barWidth; x++ {
		color := "\x1b[38;5;244m"
		var glyph byte = '-'
		if x < fill {
			color = "\x1b[38;5;45m"
			glyph = '='
		}
		setCell(grid, start+x, y, glyph, color)
	}
	text := fmt.Sprintf("SKYLINE %dk  FRAME:%06d  SAT:%02d%%", width, frame, (frame/5)%100)
	printText(grid, 2, 1, text, "\x1b[38;5;111m")
}

func updateBuildings(buildings []building, width int, frame int) {
	for i := range buildings {
		if frame%80 == 0 {
			for j := range buildings[i].windowOn {
				if rand.Intn(4) == 0 {
					buildings[i].windowOn[j] = !buildings[i].windowOn[j]
				}
			}
		}
		if rand.Intn(120) == 0 {
			buildings[i].x += 1
			if buildings[i].x > width {
				buildings[i].x = -buildings[i].width
			}
		}
	}
}

func setCell(grid [][]cell, x, y int, glyph byte, color string) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
		return
	}
	grid[y][x] = cell{glyph: glyph, color: color}
}

func printText(grid [][]cell, x, y int, text string, color string) {
	for i := 0; i < len(text); i++ {
		setCell(grid, x+i, y, text[i], color)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
