package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CribbageCui クリベッジCUIクラス
type CribbageCui struct {
	cc *controller.CribbageCuiController
}

// NewCribbageCui コンストラクタ
func NewCribbageCui() *CribbageCui {
	config := domain.DefaultCribbageConfig()
	players := []*domain.CribbagePlayer{
		domain.NewCribbagePlayer(true),
		domain.NewCribbagePlayer(false),
	}
	g := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
	return &CribbageCui{
		cc: controller.NewCribbageCuiController(usecase.NewCribbageInteractor(g, new(presenter.CribbageCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *CribbageCui) Controller() CuiExecer { return cui.cc }

// HelpLines returns the game's help lines.
func (cui *CribbageCui) HelpLines() []string {
	return []string{
		i18n.T("cribbage.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("cribbage.helpDiscard"),
		i18n.T("cribbage.helpPeg"),
		i18n.T("cribbage.helpGo"),
		i18n.T("cribbage.helpShowNext"),
		i18n.T("cribbage.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("cribbage.helpSetDifficulty"),
		i18n.T("cribbage.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *CribbageCui) Exec() {
	RunCuiLoop(cui.cc, cui.HelpLines())
}
