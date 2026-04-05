package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPokerCui コンストラクタ
func NewPokerCui() *genericCuiGame {
	config := domain.DefaultPokerConfig()
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
	pc := controller.NewPokerCuiController(usecase.NewPokerInteractor(poker, new(presenter.PokerCuiPresenter)))
	return newCuiGame(pc, []string{
		i18n.T("poker.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("poker.helpBet"),
		i18n.T("poker.helpCall"),
		i18n.T("poker.helpRaise"),
		i18n.T("poker.helpCheck"),
		i18n.T("poker.helpFold"),
		i18n.T("poker.helpAllIn"),
		i18n.T("poker.helpExchange"),
		i18n.T("poker.helpStand"),
		"",
		i18n.T("settings"),
		i18n.T("poker.helpBettingLimit"),
		i18n.T("poker.helpLowball"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
