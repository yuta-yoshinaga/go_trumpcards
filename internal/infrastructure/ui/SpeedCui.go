package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
	return newCuiGame(sc, []string{
		i18n.T("speed.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("speed.helpPlay"),
		i18n.T("speed.helpFlip"),
		i18n.T("speed.helpHint"),
		"",
		i18n.T("settings"),
		i18n.T("speed.helpSetDifficulty"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
