package ui

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// gameNames is the canonical ordered list of available game names.
var gameNames = []string{
	"blackjack", "poker", "oldmaid", "daifugo", "sevens",
	"doubt", "holdem", "hearts", "memory", "klondike", "baccarat",
}

// cuiGame is implemented by each *Cui struct to expose its controller and help lines.
type cuiGame interface {
	Controller() CuiExecer
	HelpLines() []string
}

// GameManager manages multiple game CUI controllers and enables dynamic switching.
type GameManager struct {
	games       map[string]CuiExecer
	helpLines   map[string][]string
	initialized map[string]bool
	currentGame string
	gameOrder   []string
}

// NewGameManager creates a GameManager starting with startGame (must be a valid game name).
func NewGameManager(startGame string) *GameManager {
	controllers, helpLines := buildGameEntries()
	if _, ok := controllers[startGame]; !ok {
		panic(fmt.Sprintf("NewGameManager: unknown start game %q", startGame))
	}
	return &GameManager{
		games:       controllers,
		helpLines:   helpLines,
		initialized: make(map[string]bool),
		currentGame: startGame,
		gameOrder:   gameNames,
	}
}

// Exec processes a command. "switch <game>" and "games" are handled by the manager;
// all other commands are delegated to the current game's controller.
func (m *GameManager) Exec(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) > 0 {
		switch fields[0] {
		case "switch":
			if len(fields) < 2 {
				return i18n.T("switchUsage")
			}
			return m.switchGame(fields[1])
		case "games":
			return m.listGames()
		case "help", "?":
			return strings.Join(m.HelpLines(), "\n")
		}
	}
	return m.games[m.currentGame].Exec(cmd)
}

// HelpLines returns the current game's help lines plus interactive-mode commands.
func (m *GameManager) HelpLines() []string {
	base := m.helpLines[m.currentGame]
	extra := []string{
		"",
		i18n.Tf("interactiveMode", "name", m.currentGame),
		i18n.T("switchCmd"),
		i18n.T("gamesCmd"),
	}
	lines := make([]string, len(base)+len(extra))
	copy(lines, base)
	copy(lines[len(base):], extra)
	return lines
}

// CurrentGame returns the name of the currently active game.
func (m *GameManager) CurrentGame() string {
	return m.currentGame
}

// InitCurrentGame initializes (resets) the current game if not yet done and returns the reset output.
// This should be called once at startup before entering the game loop.
func (m *GameManager) InitCurrentGame() string {
	return m.initGame(m.currentGame)
}

func (m *GameManager) initGame(name string) string {
	if !m.initialized[name] {
		m.initialized[name] = true
		return m.games[name].Exec("r")
	}
	return ""
}

func (m *GameManager) switchGame(name string) string {
	name = strings.ToLower(name)
	if _, ok := m.games[name]; !ok {
		return i18n.Tf("unknownGame", "name", name)
	}
	if name == m.currentGame {
		return i18n.Tf("alreadyPlaying", "name", name)
	}
	m.currentGame = name
	initMsg := m.initGame(name)
	msg := i18n.Tf("switchedTo", "name", name)
	if initMsg != "" {
		return msg + "\n" + initMsg
	}
	return msg
}

func (m *GameManager) listGames() string {
	var sb strings.Builder
	sb.WriteString(i18n.T("availableGames") + "\n")
	for _, name := range m.gameOrder {
		if name == m.currentGame {
			fmt.Fprintf(&sb, "  * %s %s\n", name, i18n.T("currentGame"))
		} else {
			fmt.Fprintf(&sb, "    %s\n", name)
		}
	}
	sb.WriteString(i18n.T("useSwitchCmd"))
	return sb.String()
}

func buildGameEntries() (map[string]CuiExecer, map[string][]string) {
	entries := map[string]cuiGame{
		"blackjack": NewBlackJackCui(),
		"poker":     NewPokerCui(),
		"oldmaid":   NewOldMaidCui(),
		"daifugo":   NewDaifugoCui(),
		"sevens":    NewSevensCui(),
		"doubt":     NewDoubtCui(),
		"holdem":    NewHoldemCui(),
		"hearts":    NewHeartsCui(),
		"memory":    NewMemoryCui(),
		"klondike":  NewKlondikeCui(),
		"baccarat":  NewBaccaratCui(),
	}
	controllers := make(map[string]CuiExecer, len(entries))
	helpLines := make(map[string][]string, len(entries))
	for name, g := range entries {
		controllers[name] = g.Controller()
		helpLines[name] = g.HelpLines()
	}
	return controllers, helpLines
}
