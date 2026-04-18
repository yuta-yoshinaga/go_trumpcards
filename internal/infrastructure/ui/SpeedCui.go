package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpeedCui コンストラクタ
func NewSpeedCui() *genericCuiGame {
	sc := controller.NewSpeedCuiController(
		usecase.NewSpeedInteractor(domain.NewDefaultSpeed(), new(presenter.SpeedCuiPresenter)),
	)
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "speed.helpTitle",
		CommandKeys: []string{"speed.helpPlay", "speed.helpFlip", "speed.helpHint"},
		SettingKeys: []string{"speed.helpSetDifficulty"},
	}))
}
