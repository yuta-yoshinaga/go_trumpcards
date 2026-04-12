package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCaribbeanStudCui コンストラクタ
func NewCaribbeanStudCui() *genericCuiGame {
	cs := controller.NewCaribbeanStudCuiController(usecase.NewCaribbeanStudInteractor(
		domain.NewDefaultCaribbeanStud(),
		new(presenter.CaribbeanStudCuiPresenter),
	))
	return newCuiGame(cs, []string{
		i18n.T("caribbeanstud.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("caribbeanstud.helpBet"),
		i18n.T("caribbeanstud.helpPlay"),
		i18n.T("caribbeanstud.helpFold"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
