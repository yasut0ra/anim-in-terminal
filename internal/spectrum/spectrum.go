package spectrum

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

const (
	minWidthSpectrum  = 60
	minHeightSpectrum = 24
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	barPalette = []string{
		"\x1b[38;5;33m",
		"\x1b[38;5;39m",
		"\x1b[38;5;45m",
		"\x1b[38;5;75m",
		"\x1b[38;5;111m",
	}
	tracePalette = []string{
		"\x1b[38;5;214m",
		"\x1b[38;5;221m",
		"\x1b[38;5;223m",
	}
	gridColor   = "\x1b[38;5;237m"
	peakColor   = "\x1b[38;5;229m"
	beamPalette = []string{
		"\x1b[38;5;36m",
		"\x1b[38;5;44m",
		"\x1b[38;5;51m",
	}
)

// Config controls the spectrum animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a preset tuned for a faux-equalizer view.
func DefaultConfig() Config {
	return Config{
		Width:      100,
		Height:     34,
		FrameDelay: 45 * time.Millisecond,
	}
}

func (c Config) normalize() Config {
	if c.Width < minWidthSpectrum {
		c.Width = minWidthSpectrum
	}
	if c.Height < minHeightSpectrum {
		c.Height = minHeightSpectrum
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

type bar struct {
	phase      float64
	speed      float64
	offset     float64
	colorShift int
	peak       float64
}

// Run launches the spectrum animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	bars := makeBars(max(8, cfg.Width/3))
	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		grid := newGrid(cfg.Width, cfg.Height)
		drawGrid(grid, frame)
		drawWaveform(grid, frame)
		drawBars(grid, bars, frame)
		drawScanBeam(grid, frame)
		render(grid)
		updateBars(bars)

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

func drawGrid(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height - 1
	for x := 0; x < width; x++ {
		setIfEmpty(grid, x, base, '_', gridColor)
		if x%12 == frame%12 {
			setIfEmpty(grid, x, base-6, '.', gridColor)
		}
	}

	for y := 0; y < height; y += 6 {
		for x := 0; x < width; x += 2 {
			setIfEmpty(grid, x, y, '.', gridColor)
		}
	}
}

func drawBars(grid [][]cell, bars []bar, frame int) {
	height := len(grid)
	width := len(grid[0])
	base := height - 2
	columnWidth := max(1, width/len(bars))

	for i, b := range bars {
		amp := barAmplitude(b)
		barHeight := clampInt(int(amp*(float64(height)/1.3)), 2, height-4)
		if float64(barHeight) > bars[i].peak {
			bars[i].peak = float64(barHeight)
		}
		startX := i * columnWidth

		for x := startX; x < startX+columnWidth && x < width; x++ {
			for step := 0; step < barHeight; step++ {
				y := base - step
				if y < 0 {
					continue
				}
				color := barColor(step, barHeight, frame+b.colorShift)
				glyph := barGlyph(step, barHeight)
				setCell(grid, x, y, glyph, color)
			}
		}

		peakY := base - clampInt(int(math.Round(bars[i].peak)), 1, height-3)
		center := clampInt(startX+columnWidth/2, 0, width-1)
		setCell(grid, center, peakY, '_', peakColor)
	}
}

func drawWaveform(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	center := height / 3
	for x := 0; x < width; x++ {
		fx := float64(x)
		value := math.Sin(fx*0.11+float64(frame)*0.08) +
			0.6*math.Sin(fx*0.035+float64(frame)*0.025) +
			0.3*math.Sin(fx*0.23+float64(frame)*0.12)
		y := clampInt(center-int(value*2.3), 1, height-5)
		color := tracePalette[(x/4+frame/5)%len(tracePalette)]
		setCell(grid, x, y, '*', color)
		if y+1 < height-4 {
			setCell(grid, x, y+1, '-', color)
		}
	}
}

func drawScanBeam(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	if width == 0 {
		return
	}
	beamX := (frame / 2) % width
	for offset := -1; offset <= 1; offset++ {
		col := clampInt(beamX+offset, 0, width-1)
		color := beamPalette[(offset+len(beamPalette)+frame/8)%len(beamPalette)]
		for y := 1; y < height-2; y++ {
			glyph := byte('|')
			if (y+frame/3)%4 == 0 {
				glyph = ':'
			}
			setIfEmpty(grid, col, y, glyph, color)
		}
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

func barAmplitude(b bar) float64 {
	wave := math.Sin(b.phase) + 0.7*math.Sin(b.phase*0.5+b.offset)
	return clampFloat((wave+2.0)/2.7, 0.05, 1.0)
}

func updateBars(bars []bar) {
	for i := range bars {
		bars[i].phase += bars[i].speed
		if bars[i].phase > math.Pi*2 {
			bars[i].phase -= math.Pi * 2
		}
		bars[i].speed += (rand.Float64() - 0.5) * 0.005
		bars[i].speed = clampFloat(bars[i].speed, 0.03, 0.18)
		if bars[i].peak > 0 {
			bars[i].peak -= 0.35
			if bars[i].peak < 0 {
				bars[i].peak = 0
			}
		}
	}
}

func makeBars(count int) []bar {
	result := make([]bar, count)
	for i := range result {
		result[i] = bar{
			phase:      rand.Float64() * math.Pi * 2,
			speed:      0.05 + rand.Float64()*0.08,
			offset:     rand.Float64() * math.Pi,
			colorShift: rand.Intn(len(barPalette)),
		}
	}
	return result
}

func barColor(step int, total int, frame int) string {
	if total <= 1 {
		return barPalette[0]
	}
	ratio := float64(step) / float64(total-1)
	idx := clampInt(int(ratio*float64(len(barPalette))), 0, len(barPalette)-1)
	return barPalette[(idx+frame/12)%len(barPalette)]
}

func barGlyph(step int, total int) byte {
	ratio := float64(step) / float64(max(1, total-1))
	switch {
	case ratio < 0.2:
		return '|'
	case ratio < 0.6:
		return '#'
	default:
		return '='
	}
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

func clampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
