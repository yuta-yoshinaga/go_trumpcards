package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewTriPeaksCui コンストラクタ
func NewTriPeaksCui() *genericCuiGame {
	tc := controller.NewTriPeaksCuiController(usecase.NewTriPeaksInteractor(domain.NewDefaultTriPeaks(), new(presenter.TriPeaksCuiPresenter)))
	return newCuiGame(tc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "tripeaks.helpTitle",
		CommandKeys: []string{
			"tripeaks.helpDraw",
			"tripeaks.helpRemove",
			"tripeaks.helpGiveUp",
			"tripeaks.helpHint",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
