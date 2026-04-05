package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewJokerPokerCui コンストラクタ
func NewJokerPokerCui() *genericCuiGame {
	vc := controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
		domain.NewJokerPokerVideoPoker(),
		new(presenter.VideoPokerCuiPresenter),
	))
	return newCuiGame(vc, []string{
		i18n.T("jokerpoker.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("videopoker.helpBet"),
		i18n.T("videopoker.helpHold"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
