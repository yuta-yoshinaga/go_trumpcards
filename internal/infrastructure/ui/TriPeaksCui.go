package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewTriPeaksCui コンストラクタ
func NewTriPeaksCui() *genericCuiGame {
	triPeaks := domain.NewTriPeaks(domain.NewTrumpCards(0))
	tc := controller.NewTriPeaksCuiController(usecase.NewTriPeaksInteractor(triPeaks, new(presenter.TriPeaksCuiPresenter)))
	return newCuiGame(tc, []string{
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
	})
}
