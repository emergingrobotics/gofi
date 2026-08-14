// Command gofi manages a UniFi controller's networks, DNS, reservations,
// and clients. One binary, `gofi <area> <action>`, replacing gofips,
// gofimac, gofinet, gofidns, and gofiuser (C-GLOBAL-001). See VISION.md and
// REQUIREMENTS.md for the command tree and its rationale.
package main

import (
	"errors"
	"fmt"
	"os"
)

const (
	exitOK      = 0
	exitError   = 1
	exitUsage   = 2
	exitRefused = 3
)

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gofi:", err)
		os.Exit(codeFor(err))
	}
}

// codeFor maps an error to an exit code, per C-GLOBAL-012.
func codeFor(err error) int {
	switch {
	case errors.Is(err, errRefused):
		return exitRefused
	case errors.Is(err, errUsage):
		return exitUsage
	default:
		return exitError
	}
}
