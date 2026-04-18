package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpeedCui コンストラクタ
func NewSpeedCui() *genericCuiGame {
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	config := domain.DefaultSpeedConfig()
	game := domain.NewSpeed(domain.NewTrumpCards(0), players, config)
	sc := controller.NewSpeedCuiController(
		usecase.NewSpeedInteractor(game, new(presenter.SpeedCuiPresenter)),
	)
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "speed.helpTitle",
		CommandKeys: []string{"speed.helpPlay", "speed.helpFlip", "speed.helpHint"},
		SettingKeys: []string{"speed.helpSetDifficulty"},
	}))
}
