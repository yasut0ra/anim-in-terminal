package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"animinterminal/internal/cybercube"
	"animinterminal/internal/rain"
	"animinterminal/internal/spectrum"
)

func main() {
	mode := flag.String("mode", "cybercube", "cybercube | rain | spectrum")
	width := flag.Int("width", 0, "override character width")
	height := flag.Int("height", 0, "override character height")
	delay := flag.Duration("delay", 0, "override frame delay (e.g. 50ms)")
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "cybercube", "cube":
		cfg := cybercube.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		cybercube.Run(cfg)
	case "rain", "neonrain":
		cfg := rain.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		rain.Run(cfg)
	case "spectrum", "equalizer", "scope":
		cfg := spectrum.DefaultConfig()
		applyOverrides(&cfg.Width, &cfg.Height, &cfg.FrameDelay, width, height, delay)
		spectrum.Run(cfg)
	default:
		fmt.Printf("unknown mode %q (expected cybercube | rain | spectrum)\n", *mode)
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
