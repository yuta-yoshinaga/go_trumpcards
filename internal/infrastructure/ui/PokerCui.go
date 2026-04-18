package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
	return newCuiGame(pc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "poker.helpTitle",
		CommandKeys: []string{
			"poker.helpBet",
			"poker.helpCall",
			"poker.helpRaise",
			"poker.helpCheck",
			"poker.helpFold",
			"poker.helpAllIn",
			"poker.helpExchange",
			"poker.helpStand",
		},
		SettingKeys: []string{"poker.helpBettingLimit", "poker.helpLowball"},
	}))
}
