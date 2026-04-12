package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCanfieldCui コンストラクタ
func NewCanfieldCui() *genericCuiGame {
	canfield := domain.NewCanfield(domain.NewTrumpCards(0))
	cc := controller.NewCanfieldCuiController(usecase.NewCanfieldInteractor(canfield, new(presenter.CanfieldCuiPresenter)))
	return newCuiGame(cc, []string{
		i18n.T("canfield.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("canfield.helpDraw"),
		i18n.T("canfield.helpMove"),
		i18n.T("canfield.helpMoveWF"),
		i18n.T("canfield.helpMoveRT"),
		i18n.T("canfield.helpMoveRF"),
		i18n.T("canfield.helpMoveTF"),
		i18n.T("canfield.helpMoveTT"),
		i18n.T("canfield.helpGiveUp"),
		i18n.T("canfield.helpHint"),
		i18n.T("canfield.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
