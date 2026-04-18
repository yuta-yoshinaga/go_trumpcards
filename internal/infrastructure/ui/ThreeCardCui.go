package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewThreeCardCui コンストラクタ
func NewThreeCardCui() *genericCuiGame {
	tc := controller.NewThreeCardCuiController(usecase.NewThreeCardInteractor(
		domain.NewDefaultThreeCard(),
		new(presenter.ThreeCardCuiPresenter),
	))
	return newCuiGame(tc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "threecard.helpTitle",
		CommandKeys:       []string{"threecard.helpBet", "threecard.helpPlay", "threecard.helpFold"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
