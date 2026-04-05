package ui

// genericCuiGame eliminates per-game struct boilerplate by implementing
// cuiGame with a controller and help lines provided at construction time.
type genericCuiGame struct {
	controller CuiExecer
	helpLines  []string
}

// newCuiGame creates a genericCuiGame from a controller and help lines.
func newCuiGame(controller CuiExecer, helpLines []string) *genericCuiGame {
	return &genericCuiGame{controller: controller, helpLines: helpLines}
}

// Controller returns the game controller.
func (g *genericCuiGame) Controller() CuiExecer { return g.controller }

// HelpLines returns the game's help lines.
func (g *genericCuiGame) HelpLines() []string { return g.helpLines }

// Exec runs the CUI game loop.
func (g *genericCuiGame) Exec() { RunCuiLoop(g.controller, g.helpLines) }
