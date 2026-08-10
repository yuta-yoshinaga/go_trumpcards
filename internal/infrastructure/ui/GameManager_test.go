//go:build test

package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestMain(m *testing.M) {
	i18n.SetLang("en")
	os.Exit(m.Run())
}

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

func TestGameManager_ExecSwitchUnknownWithSuggestion(t *testing.T) {
	// Use NewGameManager with real game names so Levenshtein can find a match.
	mgr := NewGameManager("blackjack")
	res := mgr.Exec("switch pokr") // typo for "poker"
	assert.Equal(t, "blackjack", mgr.CurrentGame())
	assert.Contains(t, res, "Unknown game")
	assert.Contains(t, res, "Did you mean")
	assert.Contains(t, res, "poker")
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
	assert.Len(t, mgr.games, len(GameNames()))
}

func TestGameManager_NewGameManager_AllGamesRegistered(t *testing.T) {
	mgr := NewGameManager("blackjack")
	// Verify that recently-added games are registered.
	for _, name := range []string{"canasta", "bridge", "pineapple", "gofish", "pinochle", "pigtail", "sevencardstud", "clocksolitaire", "durak", "fortythieves", "paigow"} {
		assert.Contains(t, mgr.games, name, "game %q should be registered", name)
	}
	assert.Equal(t, len(GameNames()), len(mgr.games))
}

func TestGameManager_NewGameManager_PanicsOnInvalidGame(t *testing.T) {
	assert.Panics(t, func() {
		NewGameManager("chess")
	})
}

func TestGameManager_SwitchAlias(t *testing.T) {
	mgr := NewGameManager("blackjack")
	mgr.InitCurrentGame()

	// "gin" is an alias for "ginrummy"
	result := mgr.Exec("switch gin")
	assert.Contains(t, result, "ginrummy")
	assert.Equal(t, "ginrummy", mgr.CurrentGame())
}

func TestGameManager_SwitchAliasMultiple(t *testing.T) {
	mgr := NewGameManager("blackjack")
	mgr.InitCurrentGame()

	tests := []struct {
		alias     string
		canonical string
	}{
		{"7stud", "sevencardstud"},
		{"7cs", "sevencardstud"},
		{"clock", "clocksolitaire"},
		{"crazy8", "crazyeights"},
		{"indian", "indianpoker"},
		{"video", "videopoker"},
		{"deuces", "deuceswild"},
		{"joker", "jokerpoker"},
		{"short", "shortdeck"},
		{"6plus", "shortdeck"},
		{"3card", "threecard"},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			// Reset to blackjack so we can switch
			mgr.currentGame = "blackjack"
			result := mgr.Exec("switch " + tt.alias)
			assert.Contains(t, result, tt.canonical)
			assert.Equal(t, tt.canonical, mgr.CurrentGame())
		})
	}
}

// TestGameManager_SwitchAliasTypoSuggestion verifies that `switch <typo>` of
// a known alias surfaces the alias as a "did you mean" suggestion (issue
// #1602). Before #1602 the candidate list excluded aliases, so a typo of an
// alias either suggested a far-off canonical name or nothing at all.
func TestGameManager_SwitchAliasTypoSuggestion(t *testing.T) {
	tests := []struct {
		typo string
		want string
	}{
		{"gni", "gin"},       // distance 2 from alias "gin"
		{"7stu", "7stud"},    // distance 1 from alias "7stud"
		{"crazy9", "crazy8"}, // distance 1 from alias "crazy8"
	}
	for _, tt := range tests {
		t.Run(tt.typo, func(t *testing.T) {
			mgr := NewGameManager("blackjack")
			res := mgr.Exec("switch " + tt.typo)
			assert.Equal(t, "blackjack", mgr.CurrentGame(), "current game must not change on unknown switch")
			assert.Contains(t, res, "Unknown game")
			assert.Contains(t, res, "Did you mean")
			assert.Contains(t, res, tt.want)
		})
	}
}

func TestGameAliases_AllPointToValidGames(t *testing.T) {
	gameSet := make(map[string]bool, len(GameNames()))
	for _, name := range GameNames() {
		gameSet[name] = true
	}
	for alias, canonical := range GameAliases {
		assert.True(t, gameSet[canonical], "alias %q points to unknown game %q", alias, canonical)
	}
}

func TestGameRegistry_NamesMatchGameNames(t *testing.T) {
	registry := GameRegistry()
	assert.Equal(t, len(GameNames()), len(registry), "registry and GameNames() must have same length")
	for i, entry := range registry {
		assert.Equal(t, GameNames()[i], entry.Name, "registry[%d].Name must match GameNames()[%d]", i, i)
	}
}

func TestGameRegistry_DescriptionsMatchGameDescriptions(t *testing.T) {
	descs := GameDescriptions()
	for _, entry := range GameRegistry() {
		desc, ok := descs[entry.Name]
		assert.True(t, ok, "GameDescriptions must contain %q", entry.Name)
		assert.Equal(t, entry.Description(), desc)
	}
}

func TestGameRegistry_NoDuplicateNames(t *testing.T) {
	seen := make(map[string]bool, len(gameRegistry))
	for _, entry := range gameRegistry {
		assert.False(t, seen[entry.Name], "duplicate game name in registry: %q", entry.Name)
		seen[entry.Name] = true
	}
}

func TestGameRegistry_AllConstructorsNonNil(t *testing.T) {
	for _, entry := range gameRegistry {
		assert.NotNil(t, entry.NewCui, "NewCui must not be nil for %q", entry.Name)
	}
}

// BindCuiFor must not build the interactor until the game is actually started.
// gameRegistry is a package-level var, so an eager call would construct all 264
// games' domain state at process start rather than when one is played. See #5187.
func TestBindCuiFor_DoesNotBuildInteractorUntilNewCui(t *testing.T) {
	built := 0
	entry := BindCuiFor("probe",
		func() CuiExecer {
			built++
			return stubExecer{}
		},
		func(e CuiExecer) CuiExecer { return e },
		CuiHelpSpec{Body: []string{"help"}},
	)

	assert.Equal(t, "probe", entry.Name)
	assert.Equal(t, 0, built, "registration must not construct the interactor")

	g := entry.NewCui()
	assert.Equal(t, 1, built, "NewCui must construct it exactly once")
	assert.Equal(t, []string{"help"}, g.HelpLines())

	entry.NewCui()
	assert.Equal(t, 2, built, "each NewCui call gets a fresh game")
}

type stubExecer struct{}

func (stubExecer) Exec(string) string { return "" }
