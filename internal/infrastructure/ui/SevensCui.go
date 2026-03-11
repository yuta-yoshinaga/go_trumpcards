package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevensCui 7並べCUIクラス
type SevensCui struct {
	sgc *controller.SevensCuiController
}

// NewSevensCui コンストラクタ
func NewSevensCui() *SevensCui {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
	return &SevensCui{
		sgc: controller.NewSevensCuiController(
			usecase.NewSevensInteractor(sevens, new(presenter.SevensCuiPresenter)),
		),
	}
}

// Controller returns the game controller.
func (cui *SevensCui) Controller() CuiExecer { return cui.sgc }

// HelpLines returns the game's help lines.
func (cui *SevensCui) HelpLines() []string {
	return []string{
		i18n.T("sevens.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("sevens.helpPlay"),
		"",
		i18n.T("session"),
		"  r [tunnel] [joker=N] [strategy] [passes=N]  reset with options",
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *SevensCui) Exec() {
	RunCuiLoop(cui.sgc, cui.HelpLines())
}
