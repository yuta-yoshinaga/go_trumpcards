package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCrazyEightsCui コンストラクタ
func NewCrazyEightsCui() *genericCuiGame {
	cc := controller.NewCrazyEightsCuiController(usecase.NewCrazyEightsInteractor(domain.NewDefaultCrazyEights(), new(presenter.CrazyEightsCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "crazyeights.helpTitle",
		CommandKeys:       []string{"crazyeights.helpPlay", "crazyeights.helpDraw", "crazyeights.helpSuit", "crazyeights.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"crazyeights.helpSetDifficulty", "crazyeights.helpSetLimit"},
	}))
}
