package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerCui ポーカーCUIクラス
type PokerCui struct {
	pc *controller.PokerCuiController
}

// NewPokerCui コンストラクタ
func NewPokerCui() *PokerCui {
	config := domain.DefaultPokerConfig()
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
	return &PokerCui{
		pc: controller.NewPokerCuiController(usecase.NewPokerInteractor(poker, new(presenter.PokerCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *PokerCui) Controller() CuiExecer { return cui.pc }

// HelpLines returns the game's help lines.
func (cui *PokerCui) HelpLines() []string {
	return []string{
		"5-card Draw Poker (ポーカー)",
		"",
		"Game commands:",
		"  b [amount]           bet (e.g. b 20)",
		"  c                    call",
		"  ra [amount]          raise (e.g. ra 30)",
		"  ck                   check",
		"  f                    fold",
		"  a                    all-in",
		"  e [0-4]              exchange cards (e.g. e 0 2 4)",
		"  s                    stand (no exchange)",
		"",
		"Settings:",
		"  bl [0-2]             betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
		"  lw                   toggle 2-7 Lowball mode",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *PokerCui) Exec() {
	RunCuiLoop(cui.pc, cui.HelpLines())
}
