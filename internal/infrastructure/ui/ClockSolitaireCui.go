package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewClockSolitaireCui コンストラクタ
func NewClockSolitaireCui() *genericCuiGame {
	cs := domain.NewClockSolitaire(domain.NewTrumpCards(0))
	cc := controller.NewClockSolitaireCuiController(usecase.NewClockSolitaireInteractor(cs, new(presenter.ClockSolitaireCuiPresenter)))
	return newCuiGame(cc, []string{
		i18n.T("clocksolitaire.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("clocksolitaire.helpStep"),
		i18n.T("clocksolitaire.helpAutoPlay"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
