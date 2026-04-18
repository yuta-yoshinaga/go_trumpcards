package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFiftyOneCui コンストラクタ
func NewFiftyOneCui() *genericCuiGame {
	players := []*domain.FiftyOnePlayer{
		domain.NewFiftyOnePlayer(true),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
	}
	fo := domain.NewFiftyOne(domain.NewTrumpCards(0), players)
	foc := controller.NewFiftyOneCuiController(
		usecase.NewFiftyOneInteractor(fo, new(presenter.FiftyOneCuiPresenter)),
	)
	return newCuiGame(foc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "fiftyone.helpTitle",
		CommandKeys: []string{"fiftyone.helpPlay", "fiftyone.helpAll", "fiftyone.helpStop"},
		SettingKeys: []string{"fiftyone.helpSetDifficulty"},
	}))
}
