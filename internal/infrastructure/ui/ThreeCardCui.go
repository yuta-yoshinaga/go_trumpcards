package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewThreeCardCui コンストラクタ
func NewThreeCardCui() *genericCuiGame {
	tc := controller.NewThreeCardCuiController(usecase.NewThreeCardInteractor(
		domain.NewDefaultThreeCard(),
		new(presenter.ThreeCardCuiPresenter),
	))
	return newCuiGame(tc, []string{
		i18n.T("threecard.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("threecard.helpBet"),
		i18n.T("threecard.helpPlay"),
		i18n.T("threecard.helpFold"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
