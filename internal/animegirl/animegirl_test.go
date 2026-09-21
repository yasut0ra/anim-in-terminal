package animegirl

import (
	"math"
	"testing"
)

func TestConfigNormalize(t *testing.T) {
	cfg := (Config{}).normalize()
	if cfg.Width != minWidth || cfg.Height != minHeight {
		t.Fatalf("normalized size = %dx%d, want %dx%d", cfg.Width, cfg.Height, minWidth, minHeight)
	}
	if cfg.FrameDelay != defaultDelay {
		t.Fatalf("normalized delay = %v, want %v", cfg.FrameDelay, defaultDelay)
	}
}

func TestReferenceArtDimensionsAndDensity(t *testing.T) {
	if portraitSource.width != 112 || portraitSource.height != 50 {
		t.Fatalf("source dimensions = %dx%d, want 112x50", portraitSource.width, portraitSource.height)
	}
	if portraitSource.glyphs < 1200 {
		t.Fatalf("source has only %d visible glyphs", portraitSource.glyphs)
	}
}

func TestSideReferenceDimensionsAndDensity(t *testing.T) {
	if sideSource.width != 95 || sideSource.height != 46 {
		t.Fatalf("side source dimensions = %dx%d, want 95x46", sideSource.width, sideSource.height)
	}
	if sideSource.glyphs < 700 {
		t.Fatalf("side source has only %d visible glyphs", sideSource.glyphs)
	}
}

func TestThreeQuarterReferenceDimensionsAndDensity(t *testing.T) {
	if threeQuarterSource.width != 114 || threeQuarterSource.height != 55 {
		t.Fatalf("three-quarter source dimensions = %dx%d, want 114x55", threeQuarterSource.width, threeQuarterSource.height)
	}
	if threeQuarterSource.glyphs < 1300 {
		t.Fatalf("three-quarter source has only %d visible glyphs", threeQuarterSource.glyphs)
	}
}

func TestYawMovesBetweenBothProfiles(t *testing.T) {
	quarterCycle := int(math.Round(math.Pi / (2 * yawSpeed)))
	threeQuarterCycle := int(math.Round(3 * math.Pi / (2 * yawSpeed)))
	if got := yawForFrame(quarterCycle); got < maxYaw*0.99 {
		t.Fatalf("right profile yaw = %.3f, want near %.3f", got, maxYaw)
	}
	if got := yawForFrame(threeQuarterCycle); got > -maxYaw*0.99 {
		t.Fatalf("left profile yaw = %.3f, want near %.3f", got, -maxYaw)
	}
}

func TestFrontFrameKeepsReferenceDetail(t *testing.T) {
	g := newGrid(116, 51)
	drawFrame(g, 0)

	hair, skin, eyes := countStyles(g)
	if hair < 700 {
		t.Fatalf("front hair has only %d visible cells", hair)
	}
	if skin < 40 {
		t.Fatalf("front face has only %d visible cells", skin)
	}
	if eyes < 220 {
		t.Fatalf("front eyes have only %d visible cells", eyes)
	}
}

func TestProfileNarrowsAndHidesFarEye(t *testing.T) {
	profileFrame := int(math.Round(math.Pi / (2 * yawSpeed)))
	front := newGrid(116, 51)
	profile := newGrid(116, 51)
	drawFrame(front, 0)
	drawFrame(profile, profileFrame)

	frontWidth := portraitWidth(front)
	profileWidth := portraitWidth(profile)
	if profileWidth >= frontWidth-8 {
		t.Fatalf("profile width = %d, front width = %d; want a clear turn", profileWidth, frontWidth)
	}
	_, _, frontEyes := countStyles(front)
	_, _, profileEyes := countStyles(profile)
	if profileEyes >= frontEyes*3/4 {
		t.Fatalf("profile eyes = %d, front eyes = %d; far eye was not hidden", profileEyes, frontEyes)
	}
}

func TestThreeQuarterFrameKeepsReferenceDetail(t *testing.T) {
	frame := int(math.Round(math.Asin(threeQuarterTurn) / yawSpeed))
	g := newGrid(116, 51)
	drawFrame(g, frame)

	visible := 0
	for _, row := range g.cells {
		for _, c := range row {
			if c.level >= 30 {
				visible++
			}
		}
	}
	if visible < 900 {
		t.Fatalf("three-quarter frame retained only %d reference cells", visible)
	}
}

func TestProfileKeepsSideReferenceDetail(t *testing.T) {
	profileFrame := int(math.Round(math.Pi / (2 * yawSpeed)))
	g := newGrid(116, 51)
	drawFrame(g, profileFrame)

	visible := 0
	for _, row := range g.cells {
		for _, c := range row {
			if c.level >= 30 {
				visible++
			}
		}
	}
	if visible < 650 {
		t.Fatalf("profile retained only %d side-reference cells", visible)
	}
}

func TestMirrorGlyphs(t *testing.T) {
	pairs := map[byte]byte{'/': '\\', '\\': '/', '(': ')', ')': '(', '[': ']', ']': '[', '<': '>', '>': '<'}
	for input, want := range pairs {
		if got := mirrorGlyph(input); got != want {
			t.Errorf("mirrorGlyph(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestThreeQuarterUsesOppositeSourceOrientation(t *testing.T) {
	tests := []struct {
		yaw        float64
		wantMirror bool
	}{
		{yaw: maxYaw, wantMirror: true},
		{yaw: -maxYaw, wantMirror: false},
	}
	for _, test := range tests {
		threeQuarterMirror := mirrorThreeQuarter(test.yaw)
		if threeQuarterMirror != test.wantMirror {
			t.Errorf("yaw %.2f: three-quarter mirror = %v, want %v", test.yaw, threeQuarterMirror, test.wantMirror)
		}
	}
}

func TestMinimumViewportRetainsPortrait(t *testing.T) {
	frames := []int{0, int(math.Round(math.Asin(threeQuarterTurn) / yawSpeed)), int(math.Round(math.Pi / (2 * yawSpeed)))}
	for _, frame := range frames {
		g := newGrid(minWidth, minHeight)
		drawFrame(g, frame)

		visible := 0
		for _, row := range g.cells {
			for _, c := range row {
				if c.level >= 7 {
					visible++
				}
			}
		}
		if visible < 250 {
			t.Fatalf("minimum viewport retained only %d portrait cells at frame %d", visible, frame)
		}
	}
}

func countStyles(g *grid) (hair, skin, eyes int) {
	for _, row := range g.cells {
		for _, c := range row {
			switch {
			case containsColor(hairPalette, c.color):
				hair++
			case containsColor(skinPalette, c.color), containsColor(lipPalette, c.color):
				skin++
			case containsColor(eyePalette, c.color):
				eyes++
			}
		}
	}
	return hair, skin, eyes
}

func portraitWidth(g *grid) int {
	left, right := g.width, -1
	for _, row := range g.cells {
		for x, c := range row {
			if c.level < 7 {
				continue
			}
			left = min(left, x)
			right = max(right, x)
		}
	}
	if right < left {
		return 0
	}
	return right - left + 1
}

func containsColor(palette []string, color string) bool {
	for _, candidate := range palette {
		if candidate == color {
			return true
		}
	}
	return false
}
