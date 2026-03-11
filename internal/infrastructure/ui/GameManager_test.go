//go:build test

package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockExecer is a simple CuiExecer mock for testing.
type mockExecer struct {
	lastCmd string
	result  string
}

func (m *mockExecer) Exec(cmd string) string {
	m.lastCmd = cmd
	return m.result
}

// newGameManagerWithControllers creates a GameManager with the given controllers (for testing).
func newGameManagerWithControllers(games map[string]CuiExecer, helpLines map[string][]string, startGame string) *GameManager {
	order := make([]string, 0, len(games))
	for name := range games {
		order = append(order, name)
	}
	return &GameManager{
		games:       games,
		helpLines:   helpLines,
		initialized: make(map[string]bool),
		currentGame: startGame,
		gameOrder:   order,
	}
}

// newTestManager creates a GameManager with two mock games ("a" and "b") for testing.
func newTestManager(startGame string) (*GameManager, *mockExecer, *mockExecer) {
	ma := &mockExecer{result: "result-a"}
	mb := &mockExecer{result: "result-b"}
	games := map[string]CuiExecer{"a": ma, "b": mb}
	helpLines := map[string][]string{
		"a": {"help-a1", "help-a2"},
		"b": {"help-b1"},
	}
	mgr := newGameManagerWithControllers(games, helpLines, startGame)
	return mgr, ma, mb
}

func TestGameManager_CurrentGame(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	assert.Equal(t, "a", mgr.CurrentGame())
}

func TestGameManager_ExecDelegate(t *testing.T) {
	mgr, ma, _ := newTestManager("a")
	res := mgr.Exec("hit")
	assert.Equal(t, "result-a", res)
	assert.Equal(t, "hit", ma.lastCmd)
}

func TestGameManager_ExecEmpty(t *testing.T) {
	mgr, ma, _ := newTestManager("a")
	res := mgr.Exec("")
	assert.Equal(t, "result-a", res)
	assert.Equal(t, "", ma.lastCmd)
}

func TestGameManager_ExecSwitch(t *testing.T) {
	mgr, _, mb := newTestManager("a")
	mb.result = "reset-b"
	res := mgr.Exec("switch b")
	assert.Equal(t, "b", mgr.CurrentGame())
	assert.Contains(t, res, "Switched to b.")
	assert.Contains(t, res, "reset-b") // init message included
	assert.Equal(t, "r", mb.lastCmd)   // game was reset on first switch
}

func TestGameManager_ExecSwitchCaseInsensitive(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("switch B")
	assert.Equal(t, "b", mgr.CurrentGame())
	assert.Contains(t, res, "Switched to b.")
}

func TestGameManager_ExecSwitchSameGame(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("switch a")
	assert.Equal(t, "a", mgr.CurrentGame())
	assert.Contains(t, res, "Already playing a.")
}

func TestGameManager_ExecSwitchUnknown(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("switch unknown")
	assert.Equal(t, "a", mgr.CurrentGame())
	assert.Contains(t, res, "Unknown game")
}

func TestGameManager_ExecSwitchNoName(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("switch")
	assert.Contains(t, res, "Usage:")
	assert.Equal(t, "a", mgr.CurrentGame())
}

func TestGameManager_ExecGames(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("games")
	assert.Contains(t, res, "a")
	assert.Contains(t, res, "b")
	assert.Contains(t, res, "(current)")
	assert.Contains(t, res, "switch")
}

func TestGameManager_HelpLines(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	lines := mgr.HelpLines()
	// should contain base help lines
	assert.Contains(t, lines, "help-a1")
	assert.Contains(t, lines, "help-a2")
	// should contain current game indicator
	found := false
	for _, l := range lines {
		if strings.Contains(l, "current: a") {
			found = true
			break
		}
	}
	assert.True(t, found, "HelpLines should indicate current game")
	// should contain switch and games commands
	switchFound, gamesFound := false, false
	for _, l := range lines {
		if strings.Contains(l, "switch") {
			switchFound = true
		}
		if strings.Contains(l, "games") {
			gamesFound = true
		}
	}
	assert.True(t, switchFound, "HelpLines should contain switch command")
	assert.True(t, gamesFound, "HelpLines should contain games command")
}

func TestGameManager_ExecHelp(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("help")
	assert.Contains(t, res, "help-a1")
	assert.Contains(t, res, "help-a2")
	assert.Contains(t, res, "current: a")
	assert.Contains(t, res, "switch")
}

func TestGameManager_ExecHelpQuestion(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	res := mgr.Exec("?")
	assert.Contains(t, res, "help-a1")
	assert.Contains(t, res, "current: a")
}

func TestGameManager_HelpLinesChangeOnSwitch(t *testing.T) {
	mgr, _, _ := newTestManager("a")
	mgr.Exec("switch b")
	lines := mgr.HelpLines()
	assert.Contains(t, lines, "help-b1")
	found := false
	for _, l := range lines {
		if strings.Contains(l, "current: b") {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestGameManager_InitCurrentGame(t *testing.T) {
	mgr, ma, _ := newTestManager("a")
	ma.result = "init-a"
	res := mgr.InitCurrentGame()
	assert.Equal(t, "init-a", res)
	assert.Equal(t, "r", ma.lastCmd)
}

func TestGameManager_InitCurrentGame_OnlyOnce(t *testing.T) {
	mgr, ma, _ := newTestManager("a")
	ma.result = "init-a"
	mgr.InitCurrentGame()
	ma.result = "second"
	res := mgr.InitCurrentGame()
	// second call should not re-initialize
	assert.Equal(t, "", res)
}

func TestGameManager_SwitchPreservesState(t *testing.T) {
	mgr, ma, mb := newTestManager("a")
	// init game a
	mgr.InitCurrentGame()
	ma.lastCmd = ""
	// switch to b (initializes b)
	mb.result = "init-b"
	mgr.Exec("switch b")
	mb.lastCmd = ""
	// switch back to a (should NOT re-init)
	res := mgr.Exec("switch a")
	assert.Equal(t, "Switched to a.", res) // no init message
	assert.Equal(t, "", ma.lastCmd)        // no 'r' was sent to a
}

func TestGameManager_NewGameManager_Smoke(t *testing.T) {
	// Smoke test: NewGameManager should not panic and should default to blackjack.
	mgr := NewGameManager("blackjack")
	assert.Equal(t, "blackjack", mgr.CurrentGame())
	assert.NotNil(t, mgr.games["blackjack"])
	assert.Len(t, mgr.games, len(gameNames))
}

func TestGameManager_NewGameManager_PanicsOnInvalidGame(t *testing.T) {
	assert.Panics(t, func() {
		NewGameManager("chess")
	})
}
