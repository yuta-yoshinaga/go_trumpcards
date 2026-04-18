package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPaiGowCui コンストラクタ
func NewPaiGowCui() *genericCuiGame {
	pgc := controller.NewPaiGowCuiController(usecase.NewPaiGowInteractor(
		domain.NewDefaultPaiGow(),
		new(presenter.PaiGowCuiPresenter),
	))
	return newCuiGame(pgc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "paigow.helpTitle",
		CommandKeys:       []string{"paigow.helpBet", "paigow.helpSet"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
