package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewScorpionCui コンストラクタ
func NewScorpionCui() *genericCuiGame {
	scorpion := domain.NewScorpion(domain.NewTrumpCards(0))
	sc := controller.NewScorpionCuiController(usecase.NewScorpionInteractor(scorpion, new(presenter.ScorpionCuiPresenter)))
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "scorpion.helpTitle",
		CommandKeys: []string{
			"scorpion.helpMove",
			"scorpion.helpMoveTT",
			"scorpion.helpDeal",
			"scorpion.helpGiveUp",
			"scorpion.helpHint",
			"scorpion.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
