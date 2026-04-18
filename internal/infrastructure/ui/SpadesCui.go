package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpadesCui コンストラクタ
func NewSpadesCui() *genericCuiGame {
	sc := controller.NewSpadesCuiController(usecase.NewSpadesInteractor(domain.NewDefaultSpades(), new(presenter.SpadesCuiPresenter)))
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "spades.helpTitle",
		CommandKeys:       []string{"spades.helpBid", "spades.helpPlay", "spades.helpNext", "spades.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"spades.helpSetDifficulty", "spades.helpSetLimit"},
	}))
}
