package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemCui テキサスホールデムCUIクラス
type HoldemCui struct {
	hc *controller.HoldemCuiController
}

// NewHoldemCui コンストラクタ
func NewHoldemCui() *HoldemCui {
	cfg := domain.DefaultHoldemConfig()
	holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
	return &HoldemCui{
		hc: controller.NewHoldemCuiController(usecase.NewHoldemInteractor(holdem, new(presenter.HoldemCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *HoldemCui) Controller() CuiExecer { return cui.hc }

// HelpLines returns the game's help lines.
func (cui *HoldemCui) HelpLines() []string {
	return []string{
		"Texas Hold'em (テキサスホールデム)",
		"",
		"Game commands:",
		"  f                    fold",
		"  ck                   check",
		"  c                    call",
		"  b [amount]           bet (e.g. b 20)",
		"  ra [amount]          raise (e.g. ra 30)",
		"  a                    all-in",
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		"Settings:",
		"  bl [0-2]             betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
		"  tm [0-1]             tournament mode (0=OFF, 1=ON)",
		"  sb <amount>          small blind (>=1)",
		"  bb <amount>          big blind (>=2)",
		"  lh <hands>           blind level-up hands (>=1)",
		"  ts [4|6|9]           table size",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *HoldemCui) Exec() {
	RunCuiLoop(cui.hc, cui.HelpLines())
}
