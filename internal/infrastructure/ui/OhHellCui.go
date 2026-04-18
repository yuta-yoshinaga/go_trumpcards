package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewOhHellCui コンストラクタ
func NewOhHellCui() *genericCuiGame {
	oc := controller.NewOhHellCuiController(usecase.NewOhHellInteractor(domain.NewDefaultOhHell(), new(presenter.OhHellCuiPresenter)))
	return newCuiGame(oc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "ohhell.helpTitle",
		CommandKeys:       []string{"ohhell.helpBid", "ohhell.helpPlay", "ohhell.helpNext", "ohhell.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"ohhell.helpSetDifficulty", "ohhell.helpSetMaxHand"},
	}))
}
