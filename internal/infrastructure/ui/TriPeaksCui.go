package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TriPeaksCui トリピークスCUIクラス
type TriPeaksCui struct {
	tc *controller.TriPeaksCuiController
}

// NewTriPeaksCui コンストラクタ
func NewTriPeaksCui() *TriPeaksCui {
	triPeaks := domain.NewTriPeaks(domain.NewTrumpCards(0))
	return &TriPeaksCui{
		tc: controller.NewTriPeaksCuiController(usecase.NewTriPeaksInteractor(triPeaks, new(presenter.TriPeaksCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *TriPeaksCui) Controller() CuiExecer { return cui.tc }

// HelpLines returns the game's help lines.
func (cui *TriPeaksCui) HelpLines() []string {
	return []string{
		i18n.T("tripeaks.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("tripeaks.helpDraw"),
		i18n.T("tripeaks.helpRemove"),
		i18n.T("tripeaks.helpGiveUp"),
		i18n.T("tripeaks.helpHint"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *TriPeaksCui) Exec() {
	RunCuiLoop(cui.tc, cui.HelpLines())
}
