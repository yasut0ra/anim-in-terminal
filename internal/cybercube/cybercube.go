package cybercube

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	cameraDistance = 4.5
	aspectRatio    = 0.55
)

var (
	ansiReset = "\x1b[0m"
	ansiHide  = "\x1b[?25l"
	ansiShow  = "\x1b[?25h"
	ansiClear = "\x1b[2J"
	ansiHome  = "\x1b[H"

	edgePalette = []string{
		"\x1b[38;5;45m",
		"\x1b[38;5;81m",
		"\x1b[38;5;123m",
		"\x1b[38;5;159m",
		"\x1b[38;5;201m",
	}
	vertexGlowPalette = []string{
		"\x1b[38;5;195m",
		"\x1b[38;5;159m",
		"\x1b[38;5;123m",
		"\x1b[38;5;51m",
	}
	backgroundPalette = []string{
		"\x1b[38;5;236m",
		"\x1b[38;5;237m",
		"\x1b[38;5;238m",
	}
)

// Config exposes the knobs for the animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a ready-to-run configuration tuned for a typical terminal.
func DefaultConfig() Config {
	return Config{
		Width:      96,
		Height:     32,
		FrameDelay: 60 * time.Millisecond,
	}
}

func (c Config) normalize() Config {
	if c.Width < 48 {
		c.Width = 48
	}
	if c.Height < 24 {
		c.Height = 24
	}
	if c.FrameDelay <= 0 {
		c.FrameDelay = 60 * time.Millisecond
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

type vec3 struct {
	x, y, z float64
}

type point2D struct {
	x, y  int
	depth float64
}

var (
	cubeVertices = []vec3{
		{-1, -1, -1},
		{1, -1, -1},
		{1, 1, -1},
		{-1, 1, -1},
		{-1, -1, 1},
		{1, -1, 1},
		{1, 1, 1},
		{-1, 1, 1},
	}
	cubeEdges = [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	binaryChars = []byte{'0', '1'}
)

// Run starts the infinite cyber cube animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()

	fmt.Print(ansiHide, ansiClear)
	defer fmt.Print(ansiShow, ansiReset)

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	var (
		ax    float64
		ay    float64
		az    float64
		frame int
	)

	for {
		grid := newGrid(cfg.Width, cfg.Height)
		drawBackground(grid, frame)
		drawCube(grid, ax, ay, az, frame)

		render(grid)

		ax += 0.03
		ay += 0.02
		az += 0.017
		frame++

		<-ticker.C
	}
}

func newGrid(width, height int) [][]cell {
	grid := make([][]cell, height)
	for y := range grid {
		grid[y] = make([]cell, width)
		for x := range grid[y] {
			grid[y][x].glyph = ' '
		}
	}
	return grid
}

func drawBackground(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y++ {
		if (y+frame/2)%3 != 0 {
			continue
		}
		color := backgroundPalette[(y/3+frame/12)%len(backgroundPalette)]
		for x := 0; x < width; x++ {
			if (x+frame)%7 == 0 {
				glyph := binaryChars[(x+y+frame)&1]
				setIfEmpty(grid, x, y, glyph, color)
			} else if (x+frame/3)%11 == 0 {
				setIfEmpty(grid, x, y, '-', color)
			}
		}
	}
}

func drawCube(grid [][]cell, ax, ay, az float64, frame int) {
	width := len(grid[0])
	height := len(grid)
	scale := float64(min(width, height)) * 0.9

	rotated := make([]vec3, len(cubeVertices))
	for i, v := range cubeVertices {
		rotated[i] = rotate(v, ax, ay, az)
	}

	projected := make([]point2D, len(rotated))
	for i, v := range rotated {
		x, y, depth := project(v, scale, width, height)
		projected[i] = point2D{x: x, y: y, depth: depth}
	}

	for idx, edge := range cubeEdges {
		color := edgePalette[(idx+frame/5)%len(edgePalette)]
		drawEdge(grid, projected[edge[0]], projected[edge[1]], color)
	}

	for _, pt := range projected {
		setCell(grid, pt.x, pt.y, 'O', glowForDepth(pt.depth))
	}
}

func rotate(v vec3, ax, ay, az float64) vec3 {
	sinX, cosX := math.Sin(ax), math.Cos(ax)
	sinY, cosY := math.Sin(ay), math.Cos(ay)
	sinZ, cosZ := math.Sin(az), math.Cos(az)

	y := v.y*cosX - v.z*sinX
	z := v.y*sinX + v.z*cosX

	x := v.x*cosY + z*sinY
	z = -v.x*sinY + z*cosY

	x2 := x*cosZ - y*sinZ
	y2 := x*sinZ + y*cosZ

	return vec3{x: x2, y: y2, z: z}
}

func project(v vec3, scale float64, width, height int) (int, int, float64) {
	z := v.z + cameraDistance
	if z == 0 {
		z = 0.001
	}
	depth := scale / z
	x := int(float64(width)/2 + v.x*depth)
	y := int(float64(height)/2 - v.y*depth*aspectRatio)
	return x, y, depth
}

func drawEdge(grid [][]cell, from, to point2D, color string) {
	points := linePoints(from.x, from.y, to.x, to.y)
	for _, p := range points {
		setCell(grid, p[0], p[1], '#', color)
	}
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
	sb.Grow((width+10)*height + 8)
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

func glowForDepth(depth float64) string {
	switch {
	case depth > 14:
		return vertexGlowPalette[0]
	case depth > 10:
		return vertexGlowPalette[1]
	case depth > 7:
		return vertexGlowPalette[2]
	default:
		return vertexGlowPalette[3]
	}
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
