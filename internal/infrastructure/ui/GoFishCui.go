package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewGoFishCui コンストラクタ
func NewGoFishCui() *genericCuiGame {
	gfc := controller.NewGoFishCuiController(
		usecase.NewGoFishInteractor(domain.NewDefaultGoFish(), new(presenter.GoFishCuiPresenter)),
	)
	return newCuiGame(gfc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "gofish.helpTitle",
		CommandKeys: []string{"gofish.helpAsk"},
		SettingKeys: []string{"gofish.helpSetDifficulty"},
	}))
}
