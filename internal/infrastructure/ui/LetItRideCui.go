package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewLetItRideCui コンストラクタ
func NewLetItRideCui() *genericCuiGame {
	lc := controller.NewLetItRideCuiController(usecase.NewLetItRideInteractor(
		domain.NewDefaultLetItRide(),
		new(presenter.LetItRideCuiPresenter),
	))
	return newCuiGame(lc, []string{
		i18n.T("letitride.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("letitride.helpBet"),
		i18n.T("letitride.helpPull"),
		i18n.T("letitride.helpLetItRide"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
