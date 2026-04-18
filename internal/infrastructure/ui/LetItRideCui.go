package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewLetItRideCui コンストラクタ
func NewLetItRideCui() *genericCuiGame {
	lc := controller.NewLetItRideCuiController(usecase.NewLetItRideInteractor(
		domain.NewDefaultLetItRide(),
		new(presenter.LetItRideCuiPresenter),
	))
	return newCuiGame(lc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "letitride.helpTitle",
		CommandKeys:       []string{"letitride.helpBet", "letitride.helpPull", "letitride.helpLetItRide"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
