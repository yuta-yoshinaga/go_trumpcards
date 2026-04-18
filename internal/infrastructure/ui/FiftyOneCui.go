package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFiftyOneCui コンストラクタ
func NewFiftyOneCui() *genericCuiGame {
	foc := controller.NewFiftyOneCuiController(
		usecase.NewFiftyOneInteractor(domain.NewDefaultFiftyOne(), new(presenter.FiftyOneCuiPresenter)),
	)
	return newCuiGame(foc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "fiftyone.helpTitle",
		CommandKeys: []string{"fiftyone.helpPlay", "fiftyone.helpAll", "fiftyone.helpStop"},
		SettingKeys: []string{"fiftyone.helpSetDifficulty"},
	}))
}
