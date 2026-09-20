package orbit

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	minWidth         = 60
	minHeight        = 24
	minParticles     = 48
	coreRadiusFactor = 0.12
)

var (
	backgroundPalette = []string{
		"\x1b[38;5;236m",
		"\x1b[38;5;237m",
		"\x1b[38;5;238m",
	}
	ringPalette = []string{
		"\x1b[38;5;31m",
		"\x1b[38;5;33m",
		"\x1b[38;5;39m",
		"\x1b[38;5;45m",
	}
	arcPalette = []string{
		"\x1b[38;5;45m",
		"\x1b[38;5;51m",
		"\x1b[38;5;87m",
	}
	particlePalette = []string{
		"\x1b[38;5;195m",
		"\x1b[38;5;159m",
		"\x1b[38;5;123m",
	}
	corePalette = []string{
		"\x1b[38;5;200m",
		"\x1b[38;5;207m",
		"\x1b[38;5;213m",
		"\x1b[38;5;219m",
	}
	trailPalette = []string{
		"\x1b[38;5;111m",
		"\x1b[38;5;81m",
		"\x1b[38;5;51m",
	}
	uiPalette = []string{
		"\x1b[38;5;244m",
		"\x1b[38;5;246m",
	}
	haloPalette = []string{
		"\x1b[38;5;25m",
		"\x1b[38;5;27m",
		"\x1b[38;5;33m",
		"\x1b[38;5;39m",
	}
	beamPalette = []string{
		"\x1b[38;5;45m",
		"\x1b[38;5;51m",
	}
)

// Config controls the orbit HUD animation.
type Config struct {
	Width         int
	Height        int
	FrameDelay    time.Duration
	ParticleCount int
}

// DefaultConfig returns a preset suited for typical terminals.
func DefaultConfig() Config {
	return Config{
		Width:         100,
		Height:        34,
		FrameDelay:    40 * time.Millisecond,
		ParticleCount: 120,
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
	if c.ParticleCount < minParticles {
		c.ParticleCount = minParticles
	}
	return c
}

type cell struct {
	glyph byte
	color string
}

type particle struct {
	radius     float64
	angle      float64
	angularVel float64
	layer      int
	tilt       float64
	trail      [][2]int
}

type ring struct {
	radius float64
	speed  float64
	phase  float64
	width  float64
	wobble float64
	pulse  float64
}

// Run starts the particle orbit HUD animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()
	rand.Seed(time.Now().UnixNano())

	grid := newGrid(cfg.Width, cfg.Height)
	particles := makeParticles(cfg)
	rings := makeRings(cfg)

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		clearGrid(grid)
		drawBackground(grid, frame)
		drawBackgroundStars(grid, frame)
		drawRings(grid, rings, frame)
		drawCore(grid, frame)
		drawEnergyBeams(grid, frame)
		drawSensors(grid, frame)
		drawParticles(grid, particles, frame)
		drawDebris(grid, frame)
		drawHUD(grid, particles, frame)
		render(grid)

		updateParticles(particles)
		updateRings(rings)

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
		row := grid[y]
		for x := range row {
			row[x] = cell{glyph: ' ', color: ""}
		}
	}
}

func makeParticles(cfg Config) []particle {
	result := make([]particle, cfg.ParticleCount)
	for i := range result {
		layer := rand.Intn(3)
		result[i] = particle{
			radius:     0.35 + rand.Float64()*0.45 + float64(layer)*0.18,
			angle:      rand.Float64() * math.Pi * 2,
			angularVel: 0.006 + rand.Float64()*0.018 + float64(layer)*0.004,
			layer:      layer,
			tilt:       0.55 + rand.Float64()*0.25,
			trail:      make([][2]int, 0, 6),
		}
		if rand.Intn(2) == 0 {
			result[i].angularVel *= -1
		}
	}
	return result
}

func makeRings(cfg Config) []ring {
	return []ring{
		{radius: 0.32, speed: 0.004, width: 0.02, wobble: 0.05, pulse: 0.25},
		{radius: 0.55, speed: -0.0065, width: 0.024, wobble: 0.08, pulse: 0.35},
		{radius: 0.78, speed: 0.0035, width: 0.028, wobble: 0.06, pulse: 0.2},
	}
}

func drawBackground(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y += 2 {
		color := backgroundPalette[(y/2+frame/16)%len(backgroundPalette)]
		for x := (y + frame) % 6; x < width; x += 6 {
			setIfEmpty(grid, x, y, '.', color)
		}
	}
}

func drawBackgroundStars(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	for y := 0; y < height; y += 2 {
		for x := (y + frame/2) % 8; x < width; x += 8 {
			if ((x*31 + y*17 + frame) % 73) > 2 {
				continue
			}
			color := backgroundPalette[(x/3+y+frame/21)%len(backgroundPalette)]
			setIfEmpty(grid, x, y, '.', color)
		}
	}
}

func drawRings(grid [][]cell, rings []ring, frame int) {
	width := len(grid[0])
	height := len(grid)
	centerX := width / 2
	centerY := height / 2
	scale := float64(min(width, height)) * 0.9

	for idx, r := range rings {
		radius := r.radius * scale * (1 + r.wobble*math.Sin(float64(frame)*0.04+r.phase*2))
		thickness := r.width * scale * (1 + r.pulse*math.Sin(float64(frame)*0.05+r.phase))
		color := ringPalette[(idx+frame/10)%len(ringPalette)]
		drawRing(grid, centerX, centerY, radius, thickness, r.phase, color)
		drawRingArcs(grid, centerX, centerY, radius, thickness, frame, idx)
	}
}

func drawRing(grid [][]cell, cx, cy int, radius, thickness float64, phase float64, color string) {
	steps := int(radius * 8)
	if steps < 32 {
		steps = 32
	}
	for i := 0; i < steps; i++ {
		angle := float64(i)/float64(steps)*math.Pi*2 + phase
		x := cx + int(math.Cos(angle)*radius)
		y := cy + int(math.Sin(angle)*radius*0.6)
		setIfEmpty(grid, x, y, '-', color)
		if thickness > 1 {
			setIfEmpty(grid, x, y+1, '-', color)
		}
	}
}

func drawEllipse(grid [][]cell, cx, cy int, rx, ry float64, color string) {
	steps := int(rx * 6)
	if steps < 24 {
		steps = 24
	}
	for i := 0; i < steps; i++ {
		angle := float64(i) / float64(steps) * math.Pi * 2
		x := cx + int(math.Cos(angle)*rx)
		y := cy + int(math.Sin(angle)*ry)
		setIfEmpty(grid, x, y, '.', color)
	}
}

func drawCore(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	centerX := width / 2
	centerY := height / 2
	radius := float64(min(width, height)) * coreRadiusFactor
	pulse := 1 + 0.08*math.Sin(float64(frame)*0.1)
	radius *= pulse

	for y := -int(radius); y <= int(radius); y++ {
		for x := -int(radius * 1.4); x <= int(radius*1.4); x++ {
			dist := math.Sqrt(float64(x*x) + float64(y*y)*1.5)
			if dist > radius {
				continue
			}
			intensity := 1 - dist/radius
			color := corePalette[int(clampFloat(intensity*float64(len(corePalette)), 0, float64(len(corePalette)-1)))]
			setCell(grid, centerX+x, centerY+y, '*', color)
		}
	}
	setCell(grid, centerX, centerY, '#', "\x1b[38;5;231m")
	drawCoreHalo(grid, centerX, centerY, radius, frame)
}

func drawCoreHalo(grid [][]cell, cx, cy int, baseRadius float64, frame int) {
	for i := 0; i < len(haloPalette); i++ {
		r := baseRadius*1.1 + float64(i)*1.6
		color := haloPalette[(i+frame/14)%len(haloPalette)]
		drawEllipse(grid, cx, cy, r, r*0.62, color)
	}
}

func drawParticles(grid [][]cell, particles []particle, frame int) {
	width := len(grid[0])
	height := len(grid)
	centerX := width / 2
	centerY := height / 2
	scale := float64(min(width, height)) * 0.45

	for i := range particles {
		p := &particles[i]
		x := centerX + int(math.Cos(p.angle)*p.radius*scale)
		y := centerY + int(math.Sin(p.angle)*p.radius*scale*p.tilt)

		addTrailPoint(p, x, y)
		drawParticleTrail(grid, p)

		color := particlePalette[p.layer%len(particlePalette)]
		glyph := particleGlyph(frame, i)
		setCell(grid, x, y, glyph, color)
	}
}

func drawSensors(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	cx := width / 2
	cy := height / 2
	maxRadius := float64(min(width, height)) * 0.8

	for i := 0; i < 2; i++ {
		angle := float64(frame)*0.01 + float64(i)*math.Pi
		color := beamPalette[i%len(beamPalette)]
		drawSensorSweep(grid, cx, cy, angle, maxRadius, color)
	}
}

func drawEnergyBeams(grid [][]cell, frame int) {
	width := len(grid[0])
	height := len(grid)
	cx := width / 2
	cy := height / 2
	count := 8
	length := float64(min(width, height)) * 0.58

	for i := 0; i < count; i++ {
		base := float64(i)/float64(count)*math.Pi*2 + float64(frame)*0.006
		osc := math.Sin(float64(frame)*0.04+float64(i)) * 0.18
		angle := base + osc
		color := arcPalette[(i+frame/5)%len(arcPalette)]
		for r := length * 0.25; r < length; r += 1.2 {
			x := cx + int(math.Cos(angle)*r)
			y := cy + int(math.Sin(angle)*r*0.62)
			setIfEmpty(grid, x, y, '|', color)
			if r == length*0.25 {
				setIfEmpty(grid, x, y, '+', color)
			}
		}
	}
}

func drawSensorSweep(grid [][]cell, cx, cy int, angle float64, radius float64, color string) {
	for r := radius * 0.6; r < radius; r += 3 {
		x := cx + int(math.Cos(angle)*r)
		y := cy + int(math.Sin(angle)*r*0.6)
		setIfEmpty(grid, x, y, '/', color)
	}
	points := linePoints(cx, cy, cx+int(math.Cos(angle)*radius), cy+int(math.Sin(angle)*radius*0.6))
	for idx, pt := range points {
		if idx%3 != 0 {
			continue
		}
		setIfEmpty(grid, pt[0], pt[1], '.', color)
	}
}

func addTrailPoint(p *particle, x, y int) {
	p.trail = append(p.trail, [2]int{x, y})
	if len(p.trail) > 7 {
		p.trail = p.trail[len(p.trail)-7:]
	}
}

func drawParticleTrail(grid [][]cell, p *particle) {
	for i := 0; i < len(p.trail)-1; i++ {
		from := p.trail[i]
		to := p.trail[i+1]
		points := linePoints(from[0], from[1], to[0], to[1])
		color := trailPalette[min(i, len(trailPalette)-1)]
		for _, pt := range points {
			setIfEmpty(grid, pt[0], pt[1], '.', color)
		}
	}
}

func drawDebris(grid [][]cell, frame int) {
	height := len(grid)
	width := len(grid[0])
	cx := width / 2
	cy := height / 2
	count := width

	for i := 0; i < count; i++ {
		f := float64(i) + float64(frame)*0.8
		theta := float64(i%9)*0.7 + math.Sin(f*0.015+float64(frame)*0.004)
		r := math.Mod(f*0.12, float64(min(width, height))/2) * (0.6 + 0.25*math.Sin(float64(frame)*0.02))
		x := cx + int(math.Cos(theta)*r)
		y := cy + int(math.Sin(theta)*r*0.55)
		if x < 0 || x >= width || y < 0 || y >= height {
			continue
		}
		color := particlePalette[(i/3+frame/7)%len(particlePalette)]
		glyph := byte('.')
		if i%11 == 0 {
			glyph = '*'
		}
		setIfEmpty(grid, x, y, glyph, color)
	}
}

func particleGlyph(frame, index int) byte {
	switch (frame + index) % 3 {
	case 0:
		return 'o'
	case 1:
		return '*'
	default:
		return '+'
	}
}

func drawHUD(grid [][]cell, particles []particle, frame int) {
	width := len(grid[0])
	height := len(grid)
	centerY := height - 3
	color := uiPalette[frame/20%len(uiPalette)]

	barWidth := width / 3
	fill := int(float64(barWidth) * (0.5 + 0.5*math.Sin(float64(frame)*0.03)))
	x0 := (width - barWidth) / 2
	for x := 0; x < barWidth; x++ {
		glyph := '-'
		if x < fill {
			glyph = '='
		}
		setCell(grid, x0+x, centerY, byte(glyph), color)
	}

	text := fmt.Sprintf("particles:%03d  rings:%d  frame:%06d", len(particles), 3, frame)
	printText(grid, 2, 1, text, uiPalette[(frame/12+1)%len(uiPalette)])
}

func printText(grid [][]cell, x, y int, text string, color string) {
	for i := 0; i < len(text); i++ {
		setCell(grid, x+i, y, text[i], color)
	}
}

func updateParticles(particles []particle) {
	for i := range particles {
		p := &particles[i]
		p.angle += p.angularVel
		if p.angle > math.Pi*2 {
			p.angle -= math.Pi * 2
		} else if p.angle < 0 {
			p.angle += math.Pi * 2
		}
		noise := (rand.Float64() - 0.5) * 0.003
		tiltDrift := (rand.Float64() - 0.5) * 0.004
		p.radius = clampFloat(p.radius+noise, 0.25, 0.98)
		p.tilt = clampFloat(p.tilt+tiltDrift, 0.5, 0.85)
		p.angularVel = clampFloat(p.angularVel+(rand.Float64()-0.5)*0.0006, -0.025, 0.025)
	}
}

func updateRings(rings []ring) {
	for i := range rings {
		rings[i].phase += rings[i].speed
	}
}

func drawRingArcs(grid [][]cell, cx, cy int, radius, thickness float64, frame int, ringIdx int) {
	steps := int(radius * 6)
	if steps < 24 {
		steps = 24
	}
	glowShift := float64(frame)*0.12 + float64(ringIdx)
	color := arcPalette[(ringIdx+frame/8)%len(arcPalette)]
	for i := 0; i < steps; i++ {
		angle := float64(i)/float64(steps)*math.Pi*2 + glowShift*0.15
		mod := math.Sin(angle*4 + glowShift)
		if mod < 0.25 {
			continue
		}
		x := cx + int(math.Cos(angle)*radius)
		y := cy + int(math.Sin(angle)*radius*0.6)
		setIfEmpty(grid, x, y, '+', color)
		if thickness > 1 {
			setIfEmpty(grid, x, y+1, '.', color)
		}
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

func render(grid [][]cell) {
	var sb strings.Builder
	height := len(grid)
	if height == 0 {
		return
	}
	width := len(grid[0])
	sb.Grow((width+8)*height + 16)
	sb.WriteString(term.Home)

	for _, row := range grid {
		for _, c := range row {
			if c.color != "" {
				sb.WriteString(c.color)
			}
			if c.glyph == 0 {
				sb.WriteByte(' ')
			} else {
				sb.WriteByte(c.glyph)
			}
		}
		sb.WriteString(term.Reset)
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

func clampFloat(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
