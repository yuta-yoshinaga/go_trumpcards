package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OhHellCui オー・ヘルCUIクラス
type OhHellCui struct {
	oc *controller.OhHellCuiController
}

// NewOhHellCui コンストラクタ
func NewOhHellCui() *OhHellCui {
	config := domain.DefaultOhHellConfig()
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	ohHell := domain.NewOhHell(domain.NewTrumpCards(0), players, config)
	return &OhHellCui{
		oc: controller.NewOhHellCuiController(usecase.NewOhHellInteractor(ohHell, new(presenter.OhHellCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *OhHellCui) Controller() CuiExecer { return cui.oc }

// HelpLines returns the game's help lines.
func (cui *OhHellCui) HelpLines() []string {
	return []string{
		i18n.T("ohhell.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("ohhell.helpBid"),
		i18n.T("ohhell.helpPlay"),
		i18n.T("ohhell.helpNext"),
		i18n.T("ohhell.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("ohhell.helpSetDifficulty"),
		i18n.T("ohhell.helpSetMaxHand"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *OhHellCui) Exec() {
	RunCuiLoop(cui.oc, cui.HelpLines())
}
