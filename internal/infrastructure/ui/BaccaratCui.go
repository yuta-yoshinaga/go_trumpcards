package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBaccaratCui コンストラクタ
func NewBaccaratCui() *genericCuiGame {
	bc := controller.NewBaccaratCuiController(usecase.NewBaccaratInteractor(
		domain.NewDefaultBaccarat(),
		new(presenter.BaccaratCuiPresenter),
	))
	return newCuiGame(bc, []string{
		i18n.T("baccarat.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("baccarat.helpBet"),
		"  log                  action log",
		"  ch                   clear history",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
