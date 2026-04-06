package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *genericCuiGame {
	bjc := controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
		domain.NewDefaultBlackJack(),
		new(presenter.BlackJackCuiPresenter),
	))
	return newCuiGame(bjc, []string{
		i18n.T("blackjack.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("blackjack.helpBet"),
		i18n.T("blackjack.helpHit"),
		i18n.T("blackjack.helpStand"),
		i18n.T("blackjack.helpDouble"),
		i18n.T("blackjack.helpSplit"),
		i18n.T("blackjack.helpInsurance"),
		i18n.T("blackjack.helpDeclineInsurance"),
		"",
		i18n.T("settings"),
		i18n.T("blackjack.helpSetCpuCount"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
