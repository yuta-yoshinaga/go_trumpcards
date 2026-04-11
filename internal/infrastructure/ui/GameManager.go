package ui

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// GameRegistryEntry holds a game's name, description, and CUI constructor.
type GameRegistryEntry struct {
	Name        string
	Description string
	NewCui      func() cuiGame
}

// gameRegistry is the single source of truth for all games.
// Order determines display order in help, games list, and completion.
var gameRegistry = []GameRegistryEntry{
	{"blackjack", "BlackJack (ブラックジャック)", func() cuiGame { return NewBlackJackCui() }},
	{"poker", "5-card Draw Poker (ポーカー)", func() cuiGame { return NewPokerCui() }},
	{"oldmaid", "Old Maid (ババ抜き)", func() cuiGame { return NewOldMaidCui() }},
	{"daifugo", "Daifugo / Great Fool (大富豪)", func() cuiGame { return NewDaifugoCui() }},
	{"sevens", "Sevens (7並べ)", func() cuiGame { return NewSevensCui() }},
	{"doubt", "Doubt (ダウト)", func() cuiGame { return NewDoubtCui() }},
	{"holdem", "Texas Hold'em (テキサスホールデム)", func() cuiGame { return NewHoldemCui() }},
	{"omaha", "Omaha Hold'em (オマハホールデム)", func() cuiGame { return NewOmahaCui() }},
	{"shortdeck", "Short Deck (6+ Hold'em) (ショートデック)", func() cuiGame { return NewShortDeckCui() }},
	{"pineapple", "Pineapple Poker (パイナップルポーカー)", func() cuiGame { return NewPineappleCui() }},
	{"hearts", "Hearts (ハーツ)", func() cuiGame { return NewHeartsCui() }},
	{"memory", "Memory / Concentration (神経衰弱)", func() cuiGame { return NewMemoryCui() }},
	{"klondike", "Klondike Solitaire (ソリティア)", func() cuiGame { return NewKlondikeCui() }},
	{"freecell", "FreeCell (フリーセル)", func() cuiGame { return NewFreeCellCui() }},
	{"baccarat", "Baccarat (バカラ)", func() cuiGame { return NewBaccaratCui() }},
	{"spades", "Spades (スペード)", func() cuiGame { return NewSpadesCui() }},
	{"crazyeights", "Crazy Eights (クレイジーエイト)", func() cuiGame { return NewCrazyEightsCui() }},
	{"ginrummy", "Gin Rummy (ジンラミー)", func() cuiGame { return NewGinRummyCui() }},
	{"canasta", "Canasta (カナスタ)", func() cuiGame { return NewCanastaCui() }},
	{"spider", "Spider Solitaire (スパイダーソリティア)", func() cuiGame { return NewSpiderCui() }},
	{"napoleon", "Napoleon (ナポレオン)", func() cuiGame { return NewNapoleonCui() }},
	{"indianpoker", "Indian Poker (インディアンポーカー)", func() cuiGame { return NewIndianPokerCui() }},
	{"videopoker", "Video Poker Jacks or Better (ビデオポーカー)", func() cuiGame { return NewVideoPokerCui() }},
	{"deuceswild", "Deuces Wild (デューシーズワイルド)", func() cuiGame { return NewDeucesWildCui() }},
	{"jokerpoker", "Joker Poker (ジョーカーポーカー)", func() cuiGame { return NewJokerPokerCui() }},
	{"euchre", "Euchre (ユーカー)", func() cuiGame { return NewEuchreCui() }},
	{"pyramid", "Pyramid (ピラミッド)", func() cuiGame { return NewPyramidCui() }},
	{"tripeaks", "TriPeaks (トリピークス)", func() cuiGame { return NewTriPeaksCui() }},
	{"cribbage", "Cribbage (クリベッジ)", func() cuiGame { return NewCribbageCui() }},
	{"threecard", "Three Card Poker (スリーカードポーカー)", func() cuiGame { return NewThreeCardCui() }},
	{"ohhell", "Oh Hell (オー・ヘル)", func() cuiGame { return NewOhHellCui() }},
	{"bridge", "Contract Bridge (コントラクトブリッジ)", func() cuiGame { return NewBridgeCui() }},
	{"speed", "Speed (スピード)", func() cuiGame { return NewSpeedCui() }},
	{"gofish", "Go Fish (ゴーフィッシュ)", func() cuiGame { return NewGoFishCui() }},
	{"pinochle", "Pinochle (ピノクル)", func() cuiGame { return NewPinochleCui() }},
	{"golf", "Golf Solitaire (ゴルフ)", func() cuiGame { return NewGolfCui() }},
	{"pigtail", "Pig's Tail (ブタのしっぽ)", func() cuiGame { return NewPigsTailCui() }},
	{"sevencardstud", "Seven Card Stud (セブンカードスタッド)", func() cuiGame { return NewSevenCardStudCui() }},
	{"clocksolitaire", "Clock Solitaire (クロックソリティア)", func() cuiGame { return NewClockSolitaireCui() }},
	{"durak", "Durak / Fool (ドゥラーク)", func() cuiGame { return NewDurakCui() }},
	{"fortythieves", "Forty Thieves (フォーティシーブス)", func() cuiGame { return NewFortyThievesCui() }},
	{"paigow", "Pai Gow Poker (パイガオポーカー)", func() cuiGame { return NewPaiGowCui() }},
	{"twotenjack", "Two Ten Jack (ツーテンジャック)", func() cuiGame { return NewTwoTenJackCui() }},
	{"caribbeanstud", "Caribbean Stud Poker (カリビアンスタッドポーカー)", func() cuiGame { return NewCaribbeanStudCui() }},
	{"war", "War (戦争)", func() cuiGame { return NewWarCui() }},
}

// GameRegistry returns a copy of the game registry for external use.
func GameRegistry() []GameRegistryEntry {
	cp := make([]GameRegistryEntry, len(gameRegistry))
	copy(cp, gameRegistry)
	return cp
}

// GameNames returns the canonical ordered list of available game names,
// derived from gameRegistry. Returns a fresh copy on each call.
func GameNames() []string {
	names := make([]string, len(gameRegistry))
	for i, e := range gameRegistry {
		names[i] = e.Name
	}
	return names
}

// GameDescriptions returns a map of game names to their display descriptions,
// derived from gameRegistry. Returns a fresh copy on each call.
func GameDescriptions() map[string]string {
	m := make(map[string]string, len(gameRegistry))
	for _, e := range gameRegistry {
		m[e.Name] = e.Description
	}
	return m
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
	"csp":    "caribbeanstud",
	"stud":   "caribbeanstud",
	"40t":    "fortythieves",
	"pgp":    "paigow",
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
		gameOrder:   GameNames(),
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
	controllers := make(map[string]CuiExecer, len(gameRegistry))
	helpLines := make(map[string][]string, len(gameRegistry))
	for _, entry := range gameRegistry {
		g := entry.NewCui()
		controllers[entry.Name] = g.Controller()
		helpLines[entry.Name] = g.HelpLines()
	}
	return controllers, helpLines
}
