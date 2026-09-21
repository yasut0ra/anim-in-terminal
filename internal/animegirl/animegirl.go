// Package animegirl renders a detailed AA portrait turning its head from side to side.
package animegirl

import (
	_ "embed"
	"fmt"
	"math"
	"strings"
	"time"

	"animinterminal/internal/term"
)

const (
	minWidth               = 60
	minHeight              = 28
	maxYaw                 = 1.08
	yawSpeed               = 0.031
	threeQuarterTurn       = 0.58
	referenceDisplayHeight = 50.0
	defaultDelay           = 65 * time.Millisecond
)

var (
	backgroundPalette = []string{
		"\x1b[38;5;17m",
		"\x1b[38;5;18m",
		"\x1b[38;5;54m",
	}
	hairPalette = []string{
		"\x1b[38;5;53m",
		"\x1b[38;5;54m",
		"\x1b[38;5;90m",
		"\x1b[38;5;127m",
		"\x1b[38;5;171m",
	}
	skinPalette = []string{
		"\x1b[38;5;181m",
		"\x1b[38;5;217m",
		"\x1b[38;5;224m",
		"\x1b[38;5;230m",
	}
	eyePalette = []string{
		"\x1b[38;5;24m",
		"\x1b[38;5;45m",
		"\x1b[38;5;87m",
		"\x1b[38;5;231m",
	}
	lipPalette = []string{
		"\x1b[38;5;125m",
		"\x1b[38;5;168m",
		"\x1b[38;5;211m",
	}
)

// referenceArt is the front-facing AA used as the portrait texture. Horizontal
// rotation is produced by projecting these characters over an elliptical head.
const referenceArt = `                                          _____@@@MMMMMMMMMMMMM@@@____
                                    ___@@MMMM                      rMMMB@___
                                __@MMMr                                   MMM@__
                             _@@MM                                            MMM@_
                          _@MM                                                    MM@_
                       __BM                                                          MM&_
                     _@MM                                                              MM@_
                   _@MM                                                                  3Mh_
                  @MM                                                                      MM&
                _BM                                                                          MB_
               @M                                                                              M&
              @M                            ,                                                   MB
            ,MM               -            @                       B              ,              MM_
           ,M/              _M            ][                       3h              &              3M,
          ,M/              @/            ,M                         M[              B              3M,
         ,M/              @/             MM                         MM               A              3M,
         M/              ,B             AM]                         9MB              3h              9B
        AM               M             ]B||                         || B              9,              MA
       @M    r          Ah       @     M |[                   |     [h 3B              M          ]    M[
       Mh   [          ]B        [    A  3]                   |h    B   3B             3[          h   ]M
      AM   ,h          B        B    @/   B    ]|             A|    M     B      \      M           ,   MA
     ,M    B          ,B       @M   ]B    M,   ]B             M]   ,B      B_    ],     9|          B    M,
     AB   ][          A[      ]M[__@MMMMM@@A    M,           |MB   A&@@MMMMMM&__  M     ]B          3|   MB
     Mh   A           M      ,MMM ,M       9[   MB           BMB  ,B         MBMh  B     M           B   3M
    ]M    B           M     ,M M ,M         M,  MMA         ][ B  M            M_  ]B    M           9    M[
    M[   |[          |]    _M  B_M           M_ 9 3B       ,M  [ B               B_ MA   9|          ]|   ]B
   ,M    A           ]|   @M  ]M      -.,_    3h@]  B_     B  |&B     _,.-         B@MB  ]|           ]    M,
   AB    M           [[ _@      __@@@@@@__ >    MM,  3B_  B   AM    r __@@@@@@__     MMB |]           B    MB
   Mh    M           [_@M    _@MMMMMMMMMMMB_            B#          _MMMMMMMMMMMM@_     M@]           M    3M
  @M     M           [M    _@MMM  __@@@@__ M,                      ,M __@@@@__  MMMB_    3]           M     M[
  M[     B           ]|  MMMMM   @M MMMMMM&                          @M  MMMMMB   MMMMM  ||     ,     M     ]M
 ]M      M           ]]  ,MM/   Mh   ]MMMMMA                        A[    MMMMMM   9MM,  [h     h     M      M[
 B[      M      [     B,@@MB   AMB__@MMMMMMM[                      ]MM___MMMMMMMA   9MM&_M     ]      M      9B
 M   ,   B      9     M  9M|   MMMMMMMMMMMMMB                      AMMMMMMMMMMMMM   |M#  B     B      B   ,   M
|M   A   [h      A    ][  M[   MMMMMMMMMMMMM]                      [MMMMMMMMMMMMM   @M  [[    @      |]   B   Mr
]B   M   ]]      3\    M       \M   MMMM   M                        M[  MMMM   M/       M    ,/      Ah   M   B[
9B  |M    B       \\   3A       MB_      _M                          M_      _#/       Ah   ,B       M    M|  B]
[B  @M,   9,       X&   9[        9M@@@@M                              M@@@@#M        ]/   @M       ,B    MB  B[
]M  BM]    B         B_  M_                            <                             @B  _B         A    @MB  Mr
 M, BMM    M          3B__Mh    # # |                                      / ] ,    A/__@r         ,B    MMB ,M
 9B ][9A    A           MMB@M&                                                    @M@#MM           B    A/][ B[
  M& M M[   9[          MM_                                                          _MB          ]B   @M M @M
   M&9[ M_   M,         ]MM\                      A&__    __@r                      @MMh         ,M   _M ]B@M
    MMM_ M&   M          M MB_                       9MMMMr                       _@M,M         ,M   @M _MMM
      MM& 3B_  M,   h    3A  MB_                                               __@M  A     |   ,M  _M  @MM
         Ms MB_ M_   \    9\   MM@__                                        __@MM   @/    /   _M__@M >M
               M@M&  3&    M_     9MB@__                                __@MMr     @M    @h  @MMM
                   M_ 3B_   M&        9MM@@____                  ____@@MM         @M   _#/ _M
                     M&]Mh_   B_            9MMMM@@@&______&@@MMMMMr            _M   _@M_@M
                       9MM@M@___M@__                                        __@M___@M@MMr`

// sideReferenceArt is the supplied right-facing profile differential. The
// left-facing keyframe is generated by mirroring it at render time.
const sideReferenceArt = `                                       _______________
                              ____@@MMMMMMMMMM9MMMMMMMMM@@@___
                         ___@MMMr                           MMM@@__
                      _@@MM                                      9MMM@@___
                   _@MM                                                 MMB&_
                _@MM                                                        MM@_
              _@Mr                                    _r                       MM&_
            _@M                                      @                           3M@_
           @M                                      _M                          (    M&
         _MM                                      @/                            \    3M_
        @M                                       A/                   |          \,    MB
       @M                                       A/                    |B          9,    3B_
      AM                                       @/              ,      [M,          B      M,
     AM                                       ]B               [      MMB          3[      M
    @M                                        M               @      ,B 9,          M    \  B
   ,Mh                                       Ah              ,[      B   B   r      ]|    [ 9[
   BB                                       ,B               B      @/   M   |      |B    9  M
  |M                                        A               ]h     ,M    B   B      |M    ][ B
  BB                        A               B               M     ,M __@@MB>]B      @M|   |M [h
  Mh                       ,               ]h              Ah    _M@M    [  MB      MM|   ]M,Ah
 |M                        B               B               B    @M       A_# [     AMM|   MMBM
 [M                       ,h               B              A   _B         M  [_____A @M   @/ Mh
 BB                       [                [              M _@r        _,,,_        MM _A/ ]
 BB                       B               |r             ]&#r         ___&__        MMMM
 [B                       ]               ]h             B         _@MMMMMMMB      [h
 ]M  ,                    [               M              B       _@MM  ,MMA        M
  M  h                    h         _____@Mh|           ][      MMM    MB         ]h
  M_#                     h       @Mr  rMMM[A           |h    ,@MB    ]MB__|      M
  9M                      h     ,M   __   MBB           A      9M     BMMMM|      B
  A/                      |     M   M  M_  9B           9       M     BMMMMh      M
 ,B   ,                   ]     B        B  M           9             ]MM M        B__
 M    h                   ]     M      ,M   B           9              B__/          MM@
|B   [                          3\     [    ]r          |h                             A
Bh   M                           M&     M&,  B          ][        3 | ,               ,B
M   ]M|                            B__       M           B         r r                B
M   BMB                              MB&_____@[          M                           ,B
M   M M           ,      [                 M  B          |[                      _   B
]B  M 3B          B      ]                 B  9[          B                         M
 M& ][ 9A         |,      A                B   M,         9[                      _M
  3B_M  MB         M_     3&            [  M    M,         M                     @M
    9BM_  B_        M_     9B           9, 9     M_         B                 _@M
           MB__     3MB_    MB_          B  [     M&   ,     B&______________@M
              M@__    BMB__  MMB_        3M_9_      B_  B_    B_    rr99rr
                 9MB@@@M& MM@&@BBB&_      9BMM&_      M&_MM@___M&_
                                   MM@___  MB_           MMh  rMMr
                                        MMMMMMM`

//go:embed three_quarter.txt
var threeQuarterReferenceArt string

// Config controls the portrait animation.
type Config struct {
	Width      int
	Height     int
	FrameDelay time.Duration
}

// DefaultConfig returns a high-detail preset matching the source AA.
func DefaultConfig() Config {
	return Config{
		Width:      116,
		Height:     51,
		FrameDelay: defaultDelay,
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
		c.FrameDelay = defaultDelay
	}
	return c
}

type cell struct {
	glyph byte
	color string
	level int
}

type grid struct {
	width, height int
	cells         [][]cell
}

func newGrid(width, height int) *grid {
	g := &grid{width: width, height: height, cells: make([][]cell, height)}
	for y := range g.cells {
		g.cells[y] = make([]cell, width)
	}
	g.clear()
	return g
}

func (g *grid) clear() {
	for y := range g.cells {
		for x := range g.cells[y] {
			g.cells[y][x] = cell{glyph: ' ', level: -1}
		}
	}
}

func (g *grid) set(x, y int, glyph byte, color string, level int) {
	if x < 0 || x >= g.width || y < 0 || y >= g.height || level < g.cells[y][x].level {
		return
	}
	g.cells[y][x] = cell{glyph: glyph, color: color, level: level}
}

func (g *grid) render() {
	var sb strings.Builder
	sb.Grow((g.width+12)*g.height + 8)
	sb.WriteString(term.Home)
	for _, row := range g.cells {
		lastColor := ""
		for _, c := range row {
			if c.color != lastColor {
				if c.color == "" {
					sb.WriteString(term.Reset)
				} else {
					sb.WriteString(c.color)
				}
				lastColor = c.color
			}
			sb.WriteByte(c.glyph)
		}
		sb.WriteString(term.Reset)
		sb.WriteByte('\n')
	}
	fmt.Print(sb.String())
}

type artSource struct {
	rows          []string
	width, height int
	glyphs        int
}

var (
	portraitSource     = parseArt(referenceArt)
	threeQuarterSource = parseArt(threeQuarterReferenceArt)
	sideSource         = parseArt(sideReferenceArt)
)

func parseArt(art string) artSource {
	rows := strings.Split(strings.Trim(art, "\n"), "\n")
	source := artSource{
		rows:   rows,
		height: len(rows),
	}
	for _, row := range rows {
		source.width = max(source.width, len(row))
		for x := 0; x < len(row); x++ {
			if row[x] == ' ' {
				continue
			}
			source.glyphs++
		}
	}
	return source
}

type projection struct {
	scale   float64
	centerX float64
	top     float64
	yaw     float64
	frame   int
}

type artStyle int

const (
	styleHair artStyle = iota
	styleSkin
	styleEye
	styleLip
)

// Run starts the portrait animation and continues until interrupted.
func Run(cfg Config) {
	cfg = cfg.normalize()
	g := newGrid(cfg.Width, cfg.Height)

	cleanup := term.Start(true)
	defer cleanup()

	ticker := time.NewTicker(cfg.FrameDelay)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		drawFrame(g, frame)
		g.render()
		<-ticker.C
	}
}

func yawForFrame(frame int) float64 {
	return maxYaw * math.Sin(float64(frame)*yawSpeed)
}

func drawFrame(g *grid, frame int) {
	g.clear()
	drawBackground(g, frame)

	yaw := yawForFrame(frame)
	turn := math.Abs(yaw) / maxYaw
	mirror := yaw < 0
	threeQuarterMirror := mirrorThreeQuarter(yaw)
	projectedYaw := math.Copysign(turn*0.65, yaw)
	scale := baseArtScale(g)
	p := projection{
		scale:   scale,
		centerX: float64(g.width-1)/2 - math.Sin(projectedYaw)*float64(portraitSource.width)*0.08*scale,
		top:     (float64(g.height)-float64(portraitSource.height)*scale)/2 + math.Sin(float64(frame)*0.047)*0.45,
		yaw:     projectedYaw,
		frame:   frame,
	}

	if turn <= threeQuarterTurn {
		blend := smoothstep(clampFloat(turn/threeQuarterTurn, 0, 1))
		drawProjectedArt(g, p, blend, mirror)
		drawKeyframeArt(g, threeQuarterSource, frame, threeQuarterMirror, blend, true, threeQuarterStyleAt)
	} else {
		blend := smoothstep(clampFloat((turn-threeQuarterTurn)/(1-threeQuarterTurn), 0, 1))
		drawKeyframeArt(g, threeQuarterSource, frame, threeQuarterMirror, blend, false, threeQuarterStyleAt)
		drawKeyframeArt(g, sideSource, frame, mirror, blend, true, sideStyleAt)
	}
	if isBlinkFrame(frame) && turn < 0.18 {
		drawBlink(g, p)
	}
}

func baseArtScale(g *grid) float64 {
	return math.Min(1, math.Min(
		float64(g.width-2)/float64(portraitSource.width),
		float64(g.height-1)/referenceDisplayHeight,
	))
}

func drawBackground(g *grid, frame int) {
	for i := 0; i < max(14, g.width/5); i++ {
		direction := i%3 - 1
		x := positiveMod(i*37+frame/10*direction, g.width)
		y := positiveMod(i*i*11+i*7, max(1, g.height-2))
		phase := (frame/8 + i*3) % 23
		glyph := byte('.')
		if phase == 0 {
			glyph = '*'
		} else if phase == 1 || phase == 22 {
			glyph = '+'
		}
		g.set(x, y, glyph, backgroundPalette[i%len(backgroundPalette)], 0)
	}
}

func drawProjectedArt(g *grid, p projection, sideBlend float64, mirror bool) {
	if sideBlend >= 1 {
		return
	}
	blink := isBlinkFrame(p.frame)
	for sy, row := range portraitSource.rows {
		for sx := 0; sx < len(row); sx++ {
			glyph := row[sx]
			if glyph == ' ' {
				continue
			}
			if !remainingVisibleInWipe(sx, portraitSource.width, sideBlend, mirror) {
				continue
			}
			style := styleAt(sx, sy)
			if blink && style == styleEye {
				continue
			}

			extraDepth := featureDepth(sx, sy, style)
			x, y, visibleDepth := p.project(sx, sy, extraDepth)
			if style != styleHair && visibleDepth < -0.5*p.scale {
				continue
			}
			if style == styleHair && sy > 14 {
				progress := float64(sy-14) / float64(max(1, portraitSource.height-15))
				x += int(math.Round(math.Sin(float64(p.frame)*0.025+float64(sy)*0.29+float64(sx)*0.025) * progress))
			}

			color, baseLevel := colorAndLevel(style, visibleDepth, p.scale)
			g.set(x, y, transformGlyph(glyph, p.yaw), color, baseLevel)
		}
	}
}

func drawKeyframeArt(
	g *grid,
	source artSource,
	frame int,
	mirror bool,
	blend float64,
	incoming bool,
	styler func(int, int) artStyle,
) {
	if (incoming && blend <= 0) || (!incoming && blend >= 1) {
		return
	}
	baseScale := baseArtScale(g)
	scale := baseScale * referenceDisplayHeight / float64(source.height)
	left := (float64(g.width)-float64(source.width)*scale)/2 + math.Sin(float64(frame)*0.047)*0.25
	top := (float64(g.height)-referenceDisplayHeight*baseScale)/2 + math.Sin(float64(frame)*0.047)*0.45

	for sy, row := range source.rows {
		for sx := 0; sx < len(row); sx++ {
			glyph := row[sx]
			if glyph == ' ' {
				continue
			}
			drawX := sx
			if mirror {
				drawX = source.width - 1 - sx
				glyph = mirrorGlyph(glyph)
			}
			visible := incomingVisibleInWipe(drawX, source.width, blend, mirror)
			if !incoming {
				visible = remainingVisibleInWipe(drawX, source.width, blend, mirror)
			}
			if !visible {
				continue
			}
			x := int(math.Round(left + float64(drawX)*scale))
			y := int(math.Round(top + float64(sy)*scale))
			style := styler(sx, sy)
			color, level := colorAndLevel(style, 12*scale, scale)
			g.set(x, y, glyph, color, level+30)
		}
	}
}

func (p projection) project(sx, sy int, extraDepth float64) (int, int, float64) {
	sourceCenterX := float64(portraitSource.width-1) / 2
	localX := (float64(sx) - sourceCenterX) * p.scale
	u := (float64(sx) - sourceCenterX) / (float64(portraitSource.width) * 0.52)
	v := (float64(sy) - float64(portraitSource.height)*0.49) / (float64(portraitSource.height) * 0.72)
	curvature := math.Sqrt(max(0, 1-u*u-v*v*0.28))
	depth := (curvature*float64(portraitSource.width)*0.17 + extraDepth) * p.scale
	projectedX := localX*math.Cos(p.yaw) + depth*math.Sin(p.yaw)
	visibleDepth := depth*math.Cos(p.yaw) - localX*math.Sin(p.yaw)
	x := int(math.Round(p.centerX + projectedX))
	y := int(math.Round(p.top + float64(sy)*p.scale))
	return x, y, visibleDepth
}

func featureDepth(sx, sy int, style artStyle) float64 {
	center := float64(portraitSource.width-1) / 2
	distance := math.Abs(float64(sx)-center) / 18
	switch style {
	case styleEye:
		return max(0, 1.7-distance*0.3)
	case styleLip:
		return max(0, 3.5-distance)
	case styleSkin:
		if sy >= 36 && sy <= 43 {
			return max(0, 4.2-distance*1.4)
		}
	}
	return 0
}

func styleAt(x, y int) artStyle {
	if y >= 27 && y <= 37 && ((x >= 25 && x <= 51) || (x >= 63 && x <= 91)) {
		return styleEye
	}
	if y >= 38 && y <= 45 && x >= 43 && x <= 70 {
		return styleLip
	}
	if y < 27 || x < 25 || x > 91 || (y >= 27 && (x < 33 || x > 83)) {
		return styleHair
	}
	return styleSkin
}

func sideStyleAt(x, y int) artStyle {
	if y >= 17 && y <= 31 && x >= 58 {
		return styleEye
	}
	if y >= 29 && y <= 37 && x >= 63 {
		return styleLip
	}
	if y < 17 || x < 54 || y >= 37 {
		return styleHair
	}
	return styleSkin
}

func threeQuarterStyleAt(x, y int) artStyle {
	if y >= 28 && y <= 39 && x >= 31 && x <= 94 {
		return styleEye
	}
	if y >= 38 && y <= 47 && x >= 51 && x <= 91 {
		return styleLip
	}
	if y < 27 || x < 29 || x > 99 || y >= 47 {
		return styleHair
	}
	return styleSkin
}

func colorAndLevel(style artStyle, visibleDepth, scale float64) (string, int) {
	brightness := int(math.Round(visibleDepth / max(0.1, scale) / 10))
	switch style {
	case styleEye:
		return eyePalette[clampInt(2+brightness, 0, len(eyePalette)-1)], 70 + brightness
	case styleLip:
		return lipPalette[clampInt(1+brightness, 0, len(lipPalette)-1)], 60 + brightness
	case styleSkin:
		return skinPalette[clampInt(1+brightness, 0, len(skinPalette)-1)], 50 + brightness
	default:
		return hairPalette[clampInt(1+brightness, 0, len(hairPalette)-1)], 30 + brightness
	}
}

func isBlinkFrame(frame int) bool {
	phase := positiveMod(frame, 149)
	return phase >= 141 && phase <= 145
}

func drawBlink(g *grid, p projection) {
	for _, eye := range []struct{ left, right int }{{29, 49}, {67, 87}} {
		for sx := eye.left; sx <= eye.right; sx++ {
			x, y, visibleDepth := p.project(sx, 33, 1.5)
			if visibleDepth < -0.5*p.scale {
				continue
			}
			glyph := byte('-')
			if sx == eye.left {
				glyph = '\\'
			} else if sx == eye.right {
				glyph = '/'
			}
			g.set(x, y, glyph, eyePalette[2], 72)
		}
	}
}

func transformGlyph(glyph byte, yaw float64) byte {
	if math.Abs(yaw) < 0.42 {
		return glyph
	}
	if glyph == '_' && math.Abs(yaw) > 0.85 {
		return '-'
	}
	return glyph
}

func mirrorGlyph(glyph byte) byte {
	switch glyph {
	case '/':
		return '\\'
	case '\\':
		return '/'
	case '(':
		return ')'
	case ')':
		return '('
	case '[':
		return ']'
	case ']':
		return '['
	case '<':
		return '>'
	case '>':
		return '<'
	default:
		return glyph
	}
}

func mirrorThreeQuarter(yaw float64) bool {
	// The supplied three-quarter AA faces the opposite direction from the
	// supplied profile AA, so its mirroring convention is intentionally reversed.
	return yaw >= 0
}

func remainingVisibleInWipe(x, width int, blend float64, mirror bool) bool {
	normalized := float64(x) / float64(max(1, width-1))
	if mirror {
		return normalized >= blend
	}
	return normalized <= 1-blend
}

func incomingVisibleInWipe(x, width int, blend float64, mirror bool) bool {
	normalized := float64(x) / float64(max(1, width-1))
	if mirror {
		return normalized <= blend
	}
	return normalized >= 1-blend
}

func smoothstep(value float64) float64 {
	return value * value * (3 - 2*value)
}

func clampFloat(value, low, high float64) float64 {
	return math.Min(high, math.Max(low, value))
}

func positiveMod(value, modulus int) int {
	if modulus <= 0 {
		return 0
	}
	result := value % modulus
	if result < 0 {
		result += modulus
	}
	return result
}

func clampInt(value, low, high int) int {
	return min(high, max(low, value))
}
