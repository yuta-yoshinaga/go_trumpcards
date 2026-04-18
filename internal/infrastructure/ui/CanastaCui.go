package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCanastaCui コンストラクタ
func NewCanastaCui() *genericCuiGame {
	cc := controller.NewCanastaCuiController(usecase.NewCanastaInteractor(domain.NewDefaultCanasta(), new(presenter.CanastaCuiPresenter)))
	return newCuiGame(cc, []string{
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
	})
}
