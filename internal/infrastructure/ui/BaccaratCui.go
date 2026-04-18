package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBaccaratCui コンストラクタ
func NewBaccaratCui() *genericCuiGame {
	bc := controller.NewBaccaratCuiController(usecase.NewBaccaratInteractor(
		domain.NewDefaultBaccarat(),
		new(presenter.BaccaratCuiPresenter),
	))
	return newCuiGame(bc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "baccarat.helpTitle",
		CommandKeys: []string{"baccarat.helpBet"},
		ExtraCommandLines: []string{
			"  log                  action log",
			"  ch                   clear history",
		},
	}))
}
