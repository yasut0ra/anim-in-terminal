package term

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	Reset       = "\x1b[0m"
	HideCursor  = "\x1b[?25l"
	ShowCursor  = "\x1b[?25h"
	ClearScreen = "\x1b[2J"
	Home        = "\x1b[H"
)

// Start hides the cursor (and clears the screen if requested) and installs a SIGINT/SIGTERM
// handler to restore terminal state. The returned cleanup must be deferred by callers.
func Start(clear bool) func() {
	fmt.Print(HideCursor)
	if clear {
		fmt.Print(ClearScreen)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		Restore()
		os.Exit(1)
	}()

	return func() {
		signal.Stop(sig)
		Restore()
	}
}

// Restore shows the cursor and resets terminal attributes.
func Restore() {
	fmt.Print(ShowCursor, Reset)
}
