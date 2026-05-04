package ui

// genericCuiGame eliminates per-game struct boilerplate by implementing
// cuiGame with a controller and help lines provided at construction time.
type genericCuiGame struct {
	controller CuiExecer
	helpLines  []string
}

// newCuiGame creates a genericCuiGame from a controller and help lines.
// Package-internal: only cuiEntry calls this now. New CUI wiring must go
// through cuiEntry(ctrl, CuiHelpSpec{...}) — use CuiHelpSpec.Body when the
// help content does not fit the structured scaffold (issue #1460).
func newCuiGame(controller CuiExecer, helpLines []string) *genericCuiGame {
	return &genericCuiGame{controller: controller, helpLines: helpLines}
}

// Controller returns the game controller.
func (g *genericCuiGame) Controller() CuiExecer { return g.controller }

// HelpLines returns the game's help lines.
func (g *genericCuiGame) HelpLines() []string { return g.helpLines }
