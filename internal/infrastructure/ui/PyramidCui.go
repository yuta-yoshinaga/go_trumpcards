package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPyramidCui コンストラクタ
func NewPyramidCui() *genericCuiGame {
	pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
	pc := controller.NewPyramidCuiController(usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidCuiPresenter)))
	return newCuiGame(pc, []string{
		i18n.T("pyramid.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pyramid.helpDraw"),
		i18n.T("pyramid.helpRemoveKing"),
		i18n.T("pyramid.helpRemovePair"),
		i18n.T("pyramid.helpRemoveWaste"),
		i18n.T("pyramid.helpRemoveWasteKing"),
		i18n.T("pyramid.helpGiveUp"),
		i18n.T("pyramid.helpHint"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
