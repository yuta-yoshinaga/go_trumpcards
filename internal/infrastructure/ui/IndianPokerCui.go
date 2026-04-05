package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewIndianPokerCui コンストラクタ
func NewIndianPokerCui() *genericCuiGame {
	cfg := domain.DefaultIndianPokerConfig()
	ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
	ipc := controller.NewIndianPokerCuiController(usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerCuiPresenter)))
	return newCuiGame(ipc, []string{
		i18n.T("indianpoker.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("indianpoker.helpFold"),
		i18n.T("indianpoker.helpCheck"),
		i18n.T("indianpoker.helpCall"),
		i18n.T("indianpoker.helpBet"),
		i18n.T("indianpoker.helpRaise"),
		i18n.T("indianpoker.helpAllIn"),
		"",
		i18n.T("settings"),
		i18n.T("indianpoker.helpAnte"),
		i18n.T("indianpoker.helpBettingLimit"),
		i18n.T("indianpoker.helpMetaAI"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
