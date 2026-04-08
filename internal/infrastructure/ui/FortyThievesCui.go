package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFortyThievesCui コンストラクタ
func NewFortyThievesCui() *genericCuiGame {
	fortythieves := domain.NewFortyThieves(domain.NewTrumpCardsWithDecks(2, 0))
	fc := controller.NewFortyThievesCuiController(usecase.NewFortyThievesInteractor(fortythieves, new(presenter.FortyThievesCuiPresenter)))
	return newCuiGame(fc, []string{
		i18n.T("fortythieves.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("fortythieves.helpDraw"),
		i18n.T("fortythieves.helpMove"),
		i18n.T("fortythieves.helpMoveWF"),
		i18n.T("fortythieves.helpMoveTF"),
		i18n.T("fortythieves.helpMoveTT"),
		i18n.T("fortythieves.helpGiveUp"),
		i18n.T("fortythieves.helpHint"),
		i18n.T("fortythieves.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
