package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"animinterminal/internal/cloud"
	"animinterminal/internal/cybercube"
	"animinterminal/internal/rain"
	"animinterminal/internal/spectrum"
	"animinterminal/internal/starfield"
)

func main() {
	mode := flag.String("mode", "cybercube", "cybercube | rain | spectrum | cloud | starfield")
	width := flag.Int("width", 0, "override character width")
	height := flag.Int("height", 0, "override character height")
	delay := flag.Duration("delay", 0, "override frame delay (e.g. 50ms)")
	cubeLayout := flag.String("cube-layout", "multi", "cybercube layout: multi | single")
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "cybercube", "cube":
		cfg := cybercube.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		if cubeLayout != nil {
			applyCubeLayout(&cfg, *cubeLayout)
		}
		cybercube.Run(cfg)
	case "rain", "neonrain":
		cfg := rain.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		rain.Run(cfg)
	case "spectrum", "equalizer", "scope":
		cfg := spectrum.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		spectrum.Run(cfg)
	case "cloud", "clouds", "sky":
		cfg := cloud.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		cloud.Run(cfg)
	case "starfield", "warp", "stars":
		cfg := starfield.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		starfield.Run(cfg)
	default:
		fmt.Printf("unknown mode %q (expected cybercube | rain | spectrum | cloud | starfield)\n", *mode)
	}
}

func applyOverrides(width *int, height *int, delay *time.Duration, wOpt *int, hOpt *int, dOpt *time.Duration) {
	if wOpt != nil && *wOpt > 0 {
		*width = *wOpt
	}
	if hOpt != nil && *hOpt > 0 {
		*height = *hOpt
	}
	if dOpt != nil && *dOpt > 0 {
		*delay = *dOpt
	}
}
func applyCubeLayout(cfg *cybercube.Config, layout string) {
	switch strings.ToLower(layout) {
	case "", "multi", "default":
		// already multi
	case "single", "solo", "one":
		cfg.Instances = cybercube.SingleCubeInstances()
	default:
		fmt.Printf("unknown cube-layout %q (expected multi | single)\n", layout)
	}
}
