package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CanastaCui カナスタCUIクラス
type CanastaCui struct {
	cc *controller.CanastaCuiController
}

// NewCanastaCui コンストラクタ
func NewCanastaCui() *CanastaCui {
	config := domain.DefaultCanastaConfig()
	players := []*domain.CanastaPlayer{
		domain.NewCanastaPlayer(true),
		domain.NewCanastaPlayer(false),
	}
	canasta := domain.NewCanasta(domain.NewTrumpCardsWithDecks(2, 4), players, config)
	return &CanastaCui{
		cc: controller.NewCanastaCuiController(usecase.NewCanastaInteractor(canasta, new(presenter.CanastaCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *CanastaCui) Controller() CuiExecer { return cui.cc }

// HelpLines returns the game's help lines.
func (cui *CanastaCui) HelpLines() []string {
	return []string{
		"Canasta (カナスタ) Help",
		"",
		"Game Commands:",
		"  ds                   draw from stock",
		"  dd <idx,idx>         pick up discard pile (natural pair indices)",
		"  m <idx,idx;idx,idx>  meld (semicolon-separated groups)",
		"  sm                   skip meld phase",
		"  d <idx>              discard a card",
		"  go                   go out (requires canasta)",
		"  nr                   next round",
		"  l                    action log",
		"",
		"Settings:",
		"  sd <0-2>             set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
		"  sl <n>               set point limit",
		"",
		"Session:",
		"  r / reset            reset game",
		"  q / quit             quit",
		"  ? / help             show help",
	}
}

// Exec ゲーム実行
func (cui *CanastaCui) Exec() {
	RunCuiLoop(cui.cc, cui.HelpLines())
}
