package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GinRummyCui ジンラミーCUIクラス
type GinRummyCui struct {
	cc *controller.GinRummyCuiController
}

// NewGinRummyCui コンストラクタ
func NewGinRummyCui() *GinRummyCui {
	config := domain.DefaultGinRummyConfig()
	players := []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
	gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
	return &GinRummyCui{
		cc: controller.NewGinRummyCuiController(usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *GinRummyCui) Controller() CuiExecer { return cui.cc }

// HelpLines returns the game's help lines.
func (cui *GinRummyCui) HelpLines() []string {
	return []string{
		i18n.T("ginrummy.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("ginrummy.helpDrawStock"),
		i18n.T("ginrummy.helpDrawDiscard"),
		i18n.T("ginrummy.helpDiscard"),
		i18n.T("ginrummy.helpKnock"),
		i18n.T("ginrummy.helpLayoff"),
		i18n.T("ginrummy.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("ginrummy.helpSetDifficulty"),
		i18n.T("ginrummy.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *GinRummyCui) Exec() {
	RunCuiLoop(cui.cc, cui.HelpLines())
}
