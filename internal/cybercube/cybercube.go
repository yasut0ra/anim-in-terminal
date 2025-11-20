package cybercube

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	cameraDistance = 4.5
	aspectRatio    = 0.55
	maxFitAttempts = 10
)

var baseRotationSpeed = vec3{0.022, 0.017, 0.013}

var (
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
	faceFillPalette = []string{
		"\x1b[38;5;24m",
		"\x1b[38;5;31m",
		"\x1b[38;5;38m",
		"\x1b[38;5;44m",
		"\x1b[38;5;81m",
	}
	ghostPalette = []string{
		"\x1b[38;5;238m",
		"\x1b[38;5;239m",
		"\x1b[38;5;240m",
	}
	backdropPalette = []string{
		"\x1b[38;5;233m",
		"\x1b[38;5;234m",
		"\x1b[38;5;235m",
	}
)

// Config exposes the knobs for the animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
	Instances  []InstanceConfig
}

// InstanceConfig describes how each cube copy behaves/positions itself.
type InstanceConfig struct {
	Scale         float64
	OffsetX       float64
	OffsetY       float64
	RotationSpeed vec3
	RotationPhase vec3
}

// DefaultConfig returns a ready-to-run configuration tuned for a typical terminal.
func DefaultConfig() Config {
	return Config{
		Width:      96,
		Height:     32,
		FrameDelay: 45 * time.Millisecond,
		Instances:  MultiCubeInstances(),
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
	if len(c.Instances) == 0 {
		c.Instances = MultiCubeInstances()
	} else {
		for i := range c.Instances {
			c.Instances[i] = c.Instances[i].normalize()
		}
	}
	return c
}

func (ic InstanceConfig) normalize() InstanceConfig {
	if ic.Scale <= 0 {
		ic.Scale = 1
	}
	ic.OffsetX = clampFloat(ic.OffsetX, -0.9, 0.9)
	ic.OffsetY = clampFloat(ic.OffsetY, -0.9, 0.9)
	if ic.RotationSpeed == (vec3{}) {
		ic.RotationSpeed = baseRotationSpeed
	}
	return ic
}

// MultiCubeInstances returns the default three-instance layout used by DefaultConfig.
func MultiCubeInstances() []InstanceConfig {
	return cloneInstances(defaultInstances())
}

// SingleCubeInstances returns a centered single-cube layout.
func SingleCubeInstances() []InstanceConfig {
	return []InstanceConfig{
		defaultSingleInstance(),
	}
}

func defaultInstances() []InstanceConfig {
	return []InstanceConfig{
		{
			Scale:         0.9,
			OffsetX:       -0.55,
			OffsetY:       -0.12,
			RotationSpeed: vec3{0.019, 0.021, 0.015},
			RotationPhase: vec3{0.4, 0.1, 0.8},
		},
		{
			Scale:         1.05,
			OffsetX:       0,
			OffsetY:       0.05,
			RotationSpeed: baseRotationSpeed,
			RotationPhase: vec3{0.15, 0.05, 0},
		},
		{
			Scale:         0.78,
			OffsetX:       0.55,
			OffsetY:       -0.05,
			RotationSpeed: vec3{0.017, 0.02, 0.014},
			RotationPhase: vec3{0.7, 0.35, 0.2},
		},
	}
}

func defaultSingleInstance() InstanceConfig {
	return InstanceConfig{
		Scale:         1.1,
		RotationSpeed: baseRotationSpeed,
		RotationPhase: vec3{},
	}
}

type cell struct {
	glyph byte
	color string
	depth float64
}

type gridBuffer struct {
	width  int
	height int
	cells  [][]cell
}

func newGrid(width, height int) *gridBuffer {
	g := &gridBuffer{
		width:  width,
		height: height,
		cells:  make([][]cell, height),
	}
	for y := range g.cells {
		g.cells[y] = make([]cell, width)
	}
	g.Clear()
	return g
}

func (g *gridBuffer) Clear() {
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			g.cells[y][x].glyph = ' '
			g.cells[y][x].color = ""
			g.cells[y][x].depth = math.MaxFloat64
		}
	}
}

func (g *gridBuffer) Set(x, y int, glyph byte, color string, depth float64) {
	if y < 0 || y >= g.height {
		return
	}
	if x < 0 || x >= g.width {
		return
	}
	current := g.cells[y][x]
	if current.glyph != ' ' && depth >= current.depth {
		return
	}
	g.cells[y][x] = cell{glyph: glyph, color: color, depth: depth}
}

func (g *gridBuffer) SetIfEmpty(x, y int, glyph byte, color string) {
	if y < 0 || y >= g.height {
		return
	}
	if x < 0 || x >= g.width {
		return
	}
	if g.cells[y][x].glyph == ' ' {
		g.cells[y][x] = cell{glyph: glyph, color: color, depth: math.MaxFloat64}
	}
}

func (g *gridBuffer) Render() {
	var sb strings.Builder
	sb.Grow((g.width+10)*g.height + 8)
	sb.WriteString(term.Home)

	for _, row := range g.cells {
		for _, c := range row {
			if c.color != "" {
				sb.WriteString(c.color)
			}
			sb.WriteByte(c.glyph)
		}
		sb.WriteString(term.Reset)
		sb.WriteByte('\n')
	}

	fmt.Print(sb.String())
}

type vec3 struct {
	x, y, z float64
}

type point2D struct {
	x, y  int
	depth float64
}

type faceDef struct {
	indices [4]int
	glyph   byte
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
	cubeFaces = []faceDef{
		{indices: [4]int{0, 3, 2, 1}, glyph: '/'},
		{indices: [4]int{4, 5, 6, 7}, glyph: '\\'},
		{indices: [4]int{3, 7, 6, 2}, glyph: '-'},
		{indices: [4]int{0, 1, 5, 4}, glyph: '-'},
		{indices: [4]int{1, 2, 6, 5}, glyph: '='},
		{indices: [4]int{0, 4, 7, 3}, glyph: '='},
	}
	viewVector = vec3{0, 0, 1}
)

type cubeInstanceState struct {
	angles vec3
	cfg    InstanceConfig
}

// Run starts the infinite cyber cube animation loop.
func Run(cfg Config) {
	cfg = cfg.normalize()

	instances := make([]cubeInstanceState, len(cfg.Instances))
	for i, instCfg := range cfg.Instances {
		instances[i] = cubeInstanceState{
			angles: instCfg.RotationPhase,
			cfg:    instCfg,
		}
	}

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	grid := newGrid(cfg.Width, cfg.Height)

	for frame := 0; ; frame++ {
		grid.Clear()
		drawBackdrop(grid, frame)
		drawCubes(grid, instances, frame)

		grid.Render()

		updateInstanceRotations(instances)

		<-ticker.C
	}
}

func drawBackdrop(grid *gridBuffer, frame int) {
	height := grid.height
	width := grid.width
	for y := 0; y < height; y++ {
		if y%4 != 0 {
			continue
		}
		color := backdropPalette[(y/4+frame/30)%len(backdropPalette)]
		for x := 0; x < width; x += 2 {
			glyph := byte('.')
			if (x/2+y+frame/8)%5 == 0 {
				glyph = ':'
			}
			grid.SetIfEmpty(x, y, glyph, color)
		}
	}
}

func drawCubes(grid *gridBuffer, instances []cubeInstanceState, frame int) {
	if len(instances) == 0 {
		return
	}
	width := grid.width
	height := grid.height
	baseScale := float64(min(width, height)) * 1.25
	pulse := 0.85 + 0.15*math.Sin(float64(frame)*0.05)
	scale := baseScale * pulse

	for _, inst := range instances {
		drawCubeInstance(grid, inst, width, height, scale, frame)
	}
}

func drawCubeInstance(grid *gridBuffer, inst cubeInstanceState, width, height int, baseScale float64, frame int) {
	instanceScale := baseScale * inst.cfg.Scale
	if instanceScale <= 0 {
		return
	}

	rotated := make([]vec3, len(cubeVertices))
	for i, v := range cubeVertices {
		rotated[i] = rotate(v, inst.angles.x, inst.angles.y, inst.angles.z)
	}

	projected, fittedScale := projectToFit(rotated, width, height, instanceScale, 2)
	ghostScale := fittedScale * 1.08
	ghostProjected, _ := projectToFit(rotated, width, height, ghostScale, 1)

	offsetX, offsetY := instanceOffset(inst.cfg, width, height)
	shiftPoints(projected, offsetX, offsetY)
	shiftPoints(ghostProjected, offsetX, offsetY)

	drawGhostFrame(grid, ghostProjected, frame)
	drawFaces(grid, rotated, projected, frame)

	type edgeRender struct {
		from  point2D
		to    point2D
		color string
		depth float64
	}

	edges := make([]edgeRender, len(cubeEdges))
	for idx, edge := range cubeEdges {
		from := projected[edge[0]]
		to := projected[edge[1]]
		avgDepth := (from.depth + to.depth) * 0.5
		edges[idx] = edgeRender{
			from:  from,
			to:    to,
			color: edgeColor(idx, avgDepth, frame),
			depth: avgDepth,
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].depth > edges[j].depth
	})

	for _, edge := range edges {
		drawEdge(grid, edge.from, edge.to, edge.color)
	}

	for _, pt := range projected {
		grid.Set(pt.x, pt.y, 'O', glowForDepth(pt.depth), pt.depth-0.08)
	}
}

func instanceOffset(cfg InstanceConfig, width, height int) (int, int) {
	dx := int(float64(width) * cfg.OffsetX * 0.5)
	dy := int(float64(height) * cfg.OffsetY * 0.5)
	return dx, dy
}

func shiftPoints(points []point2D, dx, dy int) {
	for i := range points {
		points[i].x += dx
		points[i].y += dy
	}
}

func updateInstanceRotations(instances []cubeInstanceState) {
	for i := range instances {
		speed := instances[i].cfg.RotationSpeed
		instances[i].angles.x += speed.x
		instances[i].angles.y += speed.y
		instances[i].angles.z += speed.z
	}
}

func projectVertices(vertices []vec3, scale float64, width, height int) []point2D {
	projected := make([]point2D, len(vertices))
	for i, v := range vertices {
		x, y, depth := project(v, scale, width, height)
		projected[i] = point2D{x: x, y: y, depth: depth}
	}
	return projected
}

func projectToFit(vertices []vec3, width, height int, scale float64, margin int) ([]point2D, float64) {
	current := projectVertices(vertices, scale, width, height)
	if withinMargins(current, width, height, margin) {
		return current, scale
	}
	nextScale := scale
	for i := 0; i < maxFitAttempts; i++ {
		nextScale *= 0.94
		projected := projectVertices(vertices, nextScale, width, height)
		if withinMargins(projected, width, height, margin) {
			return projected, nextScale
		}
		current = projected
	}
	return current, nextScale
}

func withinMargins(points []point2D, width, height, margin int) bool {
	if margin <= 0 {
		margin = 1
	}
	for _, p := range points {
		if p.x < margin || p.x >= width-margin {
			return false
		}
		if p.y < margin || p.y >= height-margin {
			return false
		}
	}
	return true
}

func drawGhostFrame(grid *gridBuffer, projected []point2D, frame int) {
	if len(projected) == 0 {
		return
	}
	for idx, edge := range cubeEdges {
		color := ghostPalette[(idx+frame/6)%len(ghostPalette)]
		from := projected[edge[0]]
		to := projected[edge[1]]
		points := linePoints(from.x, from.y, to.x, to.y)
		for _, p := range points {
			depth := (from.depth+to.depth)*0.5 + 1.5
			grid.Set(p[0], p[1], '.', color, depth)
		}
	}
}

func drawFaces(grid *gridBuffer, rotated []vec3, projected []point2D, frame int) {
	for i, face := range cubeFaces {
		a := rotated[face.indices[0]]
		b := rotated[face.indices[1]]
		c := rotated[face.indices[2]]

		normal := cross(subtract(b, a), subtract(c, a))
		intensity := -dot(normalize(normal), viewVector)
		if intensity <= 0 {
			continue
		}

		color := shadeForFace(intensity, frame+i)
		p0 := projected[face.indices[0]]
		p1 := projected[face.indices[1]]
		p2 := projected[face.indices[2]]
		p3 := projected[face.indices[3]]

		fillTriangle(grid, p0, p1, p2, face.glyph, color)
		fillTriangle(grid, p0, p2, p3, face.glyph, color)
	}
}

func shadeForFace(intensity float64, frame int) string {
	levels := len(faceFillPalette)
	if levels == 0 {
		return ""
	}
	idx := int(clampFloat(intensity*float64(levels-1), 0, float64(levels-1)))
	offset := (frame / 24) % levels
	return faceFillPalette[(idx+offset)%levels]
}

func fillTriangle(grid *gridBuffer, a, b, c point2D, glyph byte, color string) {
	minX := max(0, min(a.x, min(b.x, c.x)))
	maxX := min(grid.width-1, max(a.x, max(b.x, c.x)))
	minY := max(0, min(a.y, min(b.y, c.y)))
	maxY := min(grid.height-1, max(a.y, max(b.y, c.y)))

	area := edgeFunction(a, b, c)
	if area == 0 {
		return
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := point2D{x: x, y: y}
			w0 := edgeFunction(b, c, p)
			w1 := edgeFunction(c, a, p)
			w2 := edgeFunction(a, b, p)

			if !sameSign(w0, w1, w2) {
				continue
			}

			w0 /= area
			w1 /= area
			w2 /= area
			depth := w0*a.depth + w1*b.depth + w2*c.depth

			grid.Set(x, y, glyph, color, depth+0.02)
		}
	}
}

func edgeFunction(a, b, c point2D) float64 {
	return float64(b.x-a.x)*float64(c.y-a.y) - float64(b.y-a.y)*float64(c.x-a.x)
}

func sameSign(values ...float64) bool {
	var hasPos, hasNeg bool
	for _, v := range values {
		if v > 0 {
			hasPos = true
		} else if v < 0 {
			hasNeg = true
		}
	}
	return !(hasPos && hasNeg)
}

func edgeColor(idx int, depth float64, frame int) string {
	if len(edgePalette) == 0 {
		return ""
	}
	closeness := clampInt(int((cameraDistance+1-depth)*3), 0, len(edgePalette)-1)
	offset := (frame / 8) % len(edgePalette)
	return edgePalette[(idx+offset+closeness)%len(edgePalette)]
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
	distance := v.z + cameraDistance
	if distance == 0 {
		distance = 0.001
	}
	scaleFactor := scale / distance
	x := int(float64(width)/2 + v.x*scaleFactor)
	y := int(float64(height)/2 - v.y*scaleFactor*aspectRatio)
	return x, y, distance
}

func drawEdge(grid *gridBuffer, from, to point2D, color string) {
	points := linePoints(from.x, from.y, to.x, to.y)
	if len(points) == 0 {
		return
	}
	glyph := edgeGlyph(to.x-from.x, to.y-from.y)
	for i, p := range points {
		var t float64
		if len(points) > 1 {
			t = float64(i) / float64(len(points)-1)
		} else {
			t = 0.5
		}
		depth := lerp(from.depth, to.depth, t) - 0.03
		if depth < 0 {
			depth = 0
		}
		grid.Set(p[0], p[1], glyph, color, depth)
	}
}

func edgeGlyph(dx, dy int) byte {
	adx := abs(dx)
	ady := abs(dy)
	switch {
	case adx > ady*2:
		return '-'
	case ady > adx*2:
		return '|'
	case dx*dy < 0:
		return '/'
	default:
		return '\\'
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

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
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

func subtract(a, b vec3) vec3 {
	return vec3{x: a.x - b.x, y: a.y - b.y, z: a.z - b.z}
}

func cross(a, b vec3) vec3 {
	return vec3{
		x: a.y*b.z - a.z*b.y,
		y: a.z*b.x - a.x*b.z,
		z: a.x*b.y - a.y*b.x,
	}
}

func dot(a, b vec3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func normalize(v vec3) vec3 {
	mag := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	if mag == 0 {
		return vec3{}
	}
	return vec3{x: v.x / mag, y: v.y / mag, z: v.z / mag}
}

func glowForDepth(depth float64) string {
	switch {
	case depth < cameraDistance-1.2:
		return vertexGlowPalette[0]
	case depth < cameraDistance-0.4:
		return vertexGlowPalette[1]
	case depth < cameraDistance+0.6:
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

func cloneInstances(src []InstanceConfig) []InstanceConfig {
	out := make([]InstanceConfig, len(src))
	copy(out, src)
	return out
}
