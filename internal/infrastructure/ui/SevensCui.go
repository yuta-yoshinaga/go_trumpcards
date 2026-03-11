package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
		"Sevens (7並べ)",
		"",
		"Game commands:",
		"  p [index]            play card (no index = pass)",
		"",
		"Settings:",
		"  r [tunnel] [joker=N] [strategy] [passes=N]  reset with options",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *SevensCui) Exec() {
	RunCuiLoop(cui.sgc, cui.HelpLines())
}
