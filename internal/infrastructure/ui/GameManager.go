package ui

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// gameNames is the canonical ordered list of available game names.
var gameNames = []string{
	"blackjack", "poker", "oldmaid", "daifugo", "sevens",
	"doubt", "holdem", "hearts", "memory", "klondike", "baccarat",
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
	return &GameManager{
		games:       buildControllers(),
		helpLines:   buildGameHelpLines(),
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
				return "Usage: switch <game>. Type 'games' for the list."
			}
			return m.switchGame(fields[1])
		case "games":
			return m.listGames()
		}
	}
	return m.games[m.currentGame].Exec(cmd)
}

// HelpLines returns the current game's help lines plus switch/games commands.
func (m *GameManager) HelpLines() []string {
	base := m.helpLines[m.currentGame]
	lines := make([]string, len(base)+3)
	copy(lines, base)
	lines[len(base)] = fmt.Sprintf("--- Current game: %s ---", m.currentGame)
	lines[len(base)+1] = "switch <game>・・・switch to another game (e.g. switch poker)"
	lines[len(base)+2] = "games・・・list all available games"
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
		return fmt.Sprintf("Unknown game: %q. Type 'games' for the list.", name)
	}
	if name == m.currentGame {
		return fmt.Sprintf("Already playing %s.", name)
	}
	m.currentGame = name
	initMsg := m.initGame(name)
	msg := fmt.Sprintf("Switched to %s.", name)
	if initMsg != "" {
		return msg + "\n" + initMsg
	}
	return msg
}

func (m *GameManager) listGames() string {
	var sb strings.Builder
	sb.WriteString("Available games:\n")
	for _, name := range m.gameOrder {
		if name == m.currentGame {
			fmt.Fprintf(&sb, "  * %s (current)\n", name)
		} else {
			fmt.Fprintf(&sb, "    %s\n", name)
		}
	}
	sb.WriteString("Use 'switch <game>' to switch.")
	return sb.String()
}

func buildControllers() map[string]CuiExecer {
	// BlackJack
	bjc := controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
		domain.NewDefaultBlackJack(),
		new(presenter.BlackJackCuiPresenter),
	))

	// Poker
	pokerCfg := domain.DefaultPokerConfig()
	pc := controller.NewPokerCuiController(usecase.NewPokerInteractor(
		domain.NewPoker(domain.NewTrumpCards(pokerCfg.JokerCount), []*domain.PokerPlayer{
			domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
			domain.NewPokerPlayer(false, domain.PokerStyleConservative),
			domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
			domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
		}, pokerCfg),
		new(presenter.PokerCuiPresenter),
	))

	// OldMaid
	omc := controller.NewOldMaidCuiController(usecase.NewOldMaidInteractor(
		domain.NewOldMaid(domain.NewTrumpCards(1), []*domain.OldMaidPlayer{
			domain.NewOldMaidPlayer(true),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
			domain.NewOldMaidPlayer(false),
		}),
		new(presenter.OldMaidCuiPresenter),
	))

	// Daifugo
	daifugoCfg := domain.DefaultDaifugoConfig()
	dgc := controller.NewDaifugoCuiController(usecase.NewDaifugoInteractor(
		domain.NewDaifugo(domain.NewTrumpCards(daifugoCfg.JokerCount), []*domain.DaifugoPlayer{
			domain.NewDaifugoPlayer(true),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
			domain.NewDaifugoPlayer(false),
		}, daifugoCfg),
		new(presenter.DaifugoCuiPresenter),
	))

	// Sevens
	sevensCfg := domain.DefaultSevensConfig()
	sgc := controller.NewSevensCuiController(usecase.NewSevensInteractor(
		domain.NewSevens(domain.NewTrumpCards(sevensCfg.JokerCount), []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}, sevensCfg),
		new(presenter.SevensCuiPresenter),
	))

	// Doubt
	dc := controller.NewDoubtCuiController(usecase.NewDoubtInteractor(
		domain.NewDoubt(domain.NewTrumpCards(0), []*domain.DoubtPlayer{
			domain.NewDoubtPlayer(true),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
			domain.NewDoubtPlayer(false),
		}),
		new(presenter.DoubtCuiPresenter),
	))

	// Holdem
	holdemCfg := domain.DefaultHoldemConfig()
	hc := controller.NewHoldemCuiController(usecase.NewHoldemInteractor(
		domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(holdemCfg.TableSize), holdemCfg),
		new(presenter.HoldemCuiPresenter),
	))

	// Hearts
	heartsCfg := domain.DefaultHeartsConfig()
	hrc := controller.NewHeartsCuiController(usecase.NewHeartsInteractor(
		domain.NewHearts(domain.NewTrumpCards(0), []*domain.HeartsPlayer{
			domain.NewHeartsPlayer(true),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
			domain.NewHeartsPlayer(false),
		}, heartsCfg),
		new(presenter.HeartsCuiPresenter),
	))

	// Memory
	memoryCfg := domain.DefaultMemoryConfig()
	mc := controller.NewMemoryCuiController(usecase.NewMemoryInteractor(
		domain.NewMemory(domain.NewTrumpCards(0), []*domain.MemoryPlayer{
			domain.NewMemoryPlayer(true),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
			domain.NewMemoryPlayer(false),
		}, memoryCfg),
		new(presenter.MemoryCuiPresenter),
	))

	// Klondike
	kc := controller.NewKlondikeCuiController(usecase.NewKlondikeInteractor(
		domain.NewKlondike(domain.NewTrumpCards(0)),
		new(presenter.KlondikeCuiPresenter),
	))

	// Baccarat
	bc := controller.NewBaccaratCuiController(usecase.NewBaccaratInteractor(
		domain.NewDefaultBaccarat(),
		new(presenter.BaccaratCuiPresenter),
	))

	return map[string]CuiExecer{
		"blackjack": bjc,
		"poker":     pc,
		"oldmaid":   omc,
		"daifugo":   dgc,
		"sevens":    sgc,
		"doubt":     dc,
		"holdem":    hc,
		"hearts":    hrc,
		"memory":    mc,
		"klondike":  kc,
		"baccarat":  bc,
	}
}

func buildGameHelpLines() map[string][]string {
	return map[string][]string{
		"blackjack": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"b N [ppBet] [t3Bet] [handCount]・・・bet (e.g. b 100 0 0 2)",
			"h・・・hit",
			"s・・・stand",
			"d・・・doubledown",
			"sp・・・split",
			"i・・・insurance",
			"di・・・decline insurance",
			"scc N・・・set CPU player count (0-3)",
		},
		"poker": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"b [amount]・・・bet (e.g. 'b 20')",
			"c・・・call",
			"ra [amount]・・・raise (e.g. 'ra 30')",
			"ck・・・check",
			"f・・・fold",
			"a・・・all-in",
			"e [0-4]・・・exchange (e.g. 'e 0 2 4' to exchange cards at index 0, 2, 4)",
			"s・・・stand (no exchange)",
			"bl [0-2]・・・betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
			"lw・・・toggle 2-7 Lowball mode",
		},
		"oldmaid": {
			"コマンドを入力してください。",
			"q・・・quit",
			"r・・・reset",
			"d・・・draw (カードを引く)",
			"s・・・shuffle (手札をシャッフル)",
			"ro [i0 i1 ...]・・・reorder (手札を並べ替え)",
			"sm [0-1]・・・set mode (0=Normal, 1=JijiNuki)",
			"sps [0-1]・・・set CPU placement strategy (0=OFF, 1=ON)",
			"sma [0-1]・・・set CPU memory AI (0=OFF, 1=ON)",
		},
		"daifugo": {
			"コマンドを入力してください。",
			"q・・・quit",
			"r・・・reset",
			"p [インデックス...]・・・カードを出す (インデックスなしでパス)",
			"sort [0-2]・・・手札ソート (0=強さ, 1=スート, 2=数字)",
			"sd [0-2]・・・CPU難易度 (0=Normal, 1=Easy, 2=Hard)",
			"sj [0-2]・・・ジョーカー枚数",
			"sr <rule> <0|1>・・・ローカルルール切替",
		},
		"sevens": {
			"コマンドを入力してください。",
			"q・・・quit",
			"r [tunnel] [joker=N] [strategy] [passes=N]・・・reset (オプションルール設定)",
			"p [インデックス]・・・カードを出す (インデックスなしでパス)",
		},
		"doubt": {
			"コマンドを入力してください。",
			"q・・・quit",
			"r・・・reset",
			"p <値> <idx...>・・・カードを出す (値=宣言値, idx=手札インデックス)",
			"d [playerIdx...]・・・ダウト",
			"s・・・スキップ（ダウトしない）",
			"sw <秒>・・・ダウト待機秒数設定 (1-60)",
			"sm <レベル>・・・CPU記憶力設定 (0=Easy, 1=Normal, 2=Hard)",
			"sp <上限>・・・ペナルティドロー上限設定 (0=無制限, >0=上限)",
		},
		"holdem": {
			"--- Commands ---",
			"q・・・quit",
			"r・・・reset",
			"f・・・fold",
			"ck・・・check",
			"c・・・call",
			"b [amount]・・・bet (e.g. 'b 20')",
			"ra [amount]・・・raise (e.g. 'ra 30')",
			"a・・・allin",
			"bl [0-2]・・・betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
			"tm [0-1]・・・tournament mode (0=OFF, 1=ON)",
			"sb [amount]・・・small blind (>=1)",
			"bb [amount]・・・big blind (>=2)",
			"lh [hands]・・・blind level-up hands (>=1)",
			"ts [4|6|9]・・・table size (4-max, 6-max, 9-max)",
			"rb・・・rebuy (accept rebuy)",
			"sr・・・skip rebuy (decline rebuy)",
			"ad・・・addon (accept addon)",
			"sa・・・skip addon (decline addon)",
			"----------------",
		},
		"hearts": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"pass <i1> <i2> <i3>・・・pass 3 cards",
			"p <i>・・・play card at index",
			"n・・・next trick",
			"nr・・・next round",
			"sd <0-2>・・・set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
			"sl <n>・・・set point limit",
			"l・・・action log",
		},
		"memory": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"f <pos>・・・flip card at position",
			"n・・・next (resolve flip)",
			"sd <0-2>・・・set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
			"l・・・action log",
		},
		"klondike": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"d・・・draw (stock → waste)",
			"m w t <col>・・・move waste → tableau",
			"m w f・・・move waste → foundation",
			"m t <col> f・・・move tableau → foundation",
			"m t <col> <idx> t <col>・・・move tableau → tableau",
			"g・・・give up",
			"h・・・hint",
			"ac・・・auto-complete",
			"l・・・action log",
		},
		"baccarat": {
			"Please enter a command.",
			"q・・・quit",
			"r・・・reset",
			"b N T・・・bet (e.g. b 100 0) T: 0=Player, 1=Banker, 2=Tie",
			"log・・・action log",
		},
	}
}
