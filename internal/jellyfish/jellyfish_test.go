package jellyfish

import (
	"math/rand"
	"testing"
	"time"
)

func TestConfigNormalize(t *testing.T) {
	cfg := (Config{}).normalize()
	if cfg.Width != minWidth || cfg.Height != minHeight {
		t.Fatalf("normalize size = %dx%d, want %dx%d", cfg.Width, cfg.Height, minWidth, minHeight)
	}
	if cfg.FrameDelay != 55*time.Millisecond {
		t.Fatalf("normalize delay = %v, want 55ms", cfg.FrameDelay)
	}
}

func TestDrawJellyIncludesBellAndTentacles(t *testing.T) {
	g := newGrid(80, 30)
	j := jelly{
		anchorX:        0.5,
		anchorY:        0.4,
		halfWidth:      14,
		capHeight:      6,
		tentacleLength: 10,
		palette:        cyanBell,
		level:          30,
	}
	drawJelly(g, j, 0)

	bell, tentacles := 0, 0
	centerY := int(j.anchorY * float64(g.height))
	for y, row := range g.cells {
		for _, c := range row {
			if c.level < 0 {
				continue
			}
			if y <= centerY+2 {
				bell++
			} else {
				tentacles++
			}
		}
	}
	if bell < 25 {
		t.Fatalf("bell has only %d visible cells", bell)
	}
	if tentacles < 20 {
		t.Fatalf("tentacles have only %d visible cells", tentacles)
	}
}

func TestSceneFitsMinimumViewport(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	s := newScene(minWidth, minHeight, rng)
	g := newGrid(minWidth, minHeight)
	drawWater(g, 50)
	s.drawBubbles(g, 50)
	for _, j := range s.jellies {
		drawJelly(g, j, 50)
	}

	visible := 0
	for _, row := range g.cells {
		for _, c := range row {
			if c.glyph != ' ' {
				visible++
			}
		}
	}
	if visible < 80 {
		t.Fatalf("minimum viewport rendered only %d cells", visible)
	}
}

func TestJellyChangesBetweenFrames(t *testing.T) {
	j := jelly{
		anchorX:        0.5,
		anchorY:        0.4,
		halfWidth:      14,
		capHeight:      6,
		tentacleLength: 10,
		driftX:         0.1,
		driftSpeed:     0.02,
		palette:        cyanBell,
		level:          30,
	}
	first := newGrid(80, 30)
	second := newGrid(80, 30)
	drawJelly(first, j, 0)
	drawJelly(second, j, 30)

	different := false
	for y := range first.cells {
		for x := range first.cells[y] {
			if first.cells[y][x].glyph != second.cells[y][x].glyph {
				different = true
				break
			}
		}
	}
	if !different {
		t.Fatal("jellyfish is identical across animation frames")
	}
}
