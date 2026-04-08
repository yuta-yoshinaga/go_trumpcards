package ui

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// GameNames is the canonical ordered list of available game names.
var GameNames = []string{
	"blackjack", "poker", "oldmaid", "daifugo", "sevens",
	"doubt", "holdem", "omaha", "shortdeck", "pineapple",
	"hearts", "memory", "klondike", "freecell", "baccarat",
	"spades", "crazyeights", "ginrummy", "canasta", "spider",
	"napoleon", "indianpoker", "videopoker", "deuceswild", "jokerpoker",
	"euchre", "pyramid", "tripeaks", "cribbage", "threecard",
	"ohhell", "bridge", "speed", "gofish", "pinochle", "golf",
	"pigtail", "sevencardstud", "clocksolitaire", "durak",
	"fortythieves",
}

// GameAliases maps short alias names to their canonical game names.
// Aliases are not shown in help or game lists.
var GameAliases = map[string]string{
	"7stud":  "sevencardstud",
	"7cs":    "sevencardstud",
	"clock":  "clocksolitaire",
	"crazy8": "crazyeights",
	"indian": "indianpoker",
	"video":  "videopoker",
	"deuces": "deuceswild",
	"joker":  "jokerpoker",
	"short":  "shortdeck",
	"6plus":  "shortdeck",
	"gin":    "ginrummy",
	"3card":  "threecard",
	"40t":    "fortythieves",
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
// i18n.SetLang must be called before NewGameManager: each game's HelpLines() is evaluated
// once at construction time, so changing the language afterwards will not update cached help text.
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
		gameOrder:   GameNames,
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
	if canonical, ok := GameAliases[name]; ok {
		name = canonical
	}
	if _, ok := m.games[name]; !ok {
		msg := i18n.Tf("unknownGame", "name", name)
		if suggestion := cuiutil.SuggestCommand(name, m.gameOrder, 2); suggestion != "" {
			msg += "\n  " + i18n.Tf("didYouMean", "name", suggestion)
		}
		return msg
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
		"blackjack":      NewBlackJackCui(),
		"poker":          NewPokerCui(),
		"oldmaid":        NewOldMaidCui(),
		"daifugo":        NewDaifugoCui(),
		"sevens":         NewSevensCui(),
		"doubt":          NewDoubtCui(),
		"holdem":         NewHoldemCui(),
		"omaha":          NewOmahaCui(),
		"shortdeck":      NewShortDeckCui(),
		"pineapple":      NewPineappleCui(),
		"hearts":         NewHeartsCui(),
		"memory":         NewMemoryCui(),
		"klondike":       NewKlondikeCui(),
		"freecell":       NewFreeCellCui(),
		"baccarat":       NewBaccaratCui(),
		"spades":         NewSpadesCui(),
		"crazyeights":    NewCrazyEightsCui(),
		"ginrummy":       NewGinRummyCui(),
		"canasta":        NewCanastaCui(),
		"spider":         NewSpiderCui(),
		"napoleon":       NewNapoleonCui(),
		"indianpoker":    NewIndianPokerCui(),
		"videopoker":     NewVideoPokerCui(),
		"deuceswild":     NewDeucesWildCui(),
		"jokerpoker":     NewJokerPokerCui(),
		"euchre":         NewEuchreCui(),
		"pyramid":        NewPyramidCui(),
		"tripeaks":       NewTriPeaksCui(),
		"cribbage":       NewCribbageCui(),
		"threecard":      NewThreeCardCui(),
		"ohhell":         NewOhHellCui(),
		"bridge":         NewBridgeCui(),
		"speed":          NewSpeedCui(),
		"gofish":         NewGoFishCui(),
		"pinochle":       NewPinochleCui(),
		"golf":           NewGolfCui(),
		"pigtail":        NewPigsTailCui(),
		"sevencardstud":  NewSevenCardStudCui(),
		"clocksolitaire": NewClockSolitaireCui(),
		"durak":          NewDurakCui(),
		"fortythieves":   NewFortyThievesCui(),
	}
	controllers := make(map[string]CuiExecer, len(entries))
	helpLines := make(map[string][]string, len(entries))
	for name, g := range entries {
		controllers[name] = g.Controller()
		helpLines[name] = g.HelpLines()
	}
	return controllers, helpLines
}
