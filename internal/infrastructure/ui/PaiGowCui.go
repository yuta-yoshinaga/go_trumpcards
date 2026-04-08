package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPaiGowCui コンストラクタ
func NewPaiGowCui() *genericCuiGame {
	pgc := controller.NewPaiGowCuiController(usecase.NewPaiGowInteractor(
		domain.NewDefaultPaiGow(),
		new(presenter.PaiGowCuiPresenter),
	))
	return newCuiGame(pgc, []string{
		i18n.T("paigow.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("paigow.helpBet"),
		i18n.T("paigow.helpSet"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
