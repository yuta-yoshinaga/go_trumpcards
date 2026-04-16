package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewRedDogCui コンストラクタ
func NewRedDogCui() *genericCuiGame {
	rc := controller.NewRedDogCuiController(usecase.NewRedDogInteractor(
		domain.NewDefaultRedDog(),
		new(presenter.RedDogCuiPresenter),
	))
	return newCuiGame(rc, []string{
		i18n.T("reddog.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("reddog.helpBet"),
		i18n.T("reddog.helpRaise"),
		i18n.T("reddog.helpStay"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
