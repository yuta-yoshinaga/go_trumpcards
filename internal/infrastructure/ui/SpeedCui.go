package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpeedCui スピードCUIクラス
type SpeedCui struct {
	sc *controller.SpeedCuiController
}

// NewSpeedCui コンストラクタ
func NewSpeedCui() *SpeedCui {
	players := []*domain.SpeedPlayer{
		domain.NewSpeedPlayer(true),
		domain.NewSpeedPlayer(false),
	}
	config := domain.DefaultSpeedConfig()
	game := domain.NewSpeed(domain.NewTrumpCards(0), players, config)
	sc := controller.NewSpeedCuiController(
		usecase.NewSpeedInteractor(game, new(presenter.SpeedCuiPresenter)),
	)
	return &SpeedCui{sc: sc}
}

// Controller returns the game controller.
func (cui *SpeedCui) Controller() CuiExecer { return cui.sc }

// HelpLines returns the game's help lines.
func (cui *SpeedCui) HelpLines() []string {
	return []string{
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
	}
}

// Exec ゲームメインループ
func (cui *SpeedCui) Exec() {
	RunCuiLoop(cui.sc, cui.HelpLines())
}
