package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewIndianPokerCui コンストラクタ
func NewIndianPokerCui() *genericCuiGame {
	cfg := domain.DefaultIndianPokerConfig()
	ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
	ipc := controller.NewIndianPokerCuiController(usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerCuiPresenter)))
	return newCuiGame(ipc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "indianpoker.helpTitle",
		CommandKeys: []string{
			"indianpoker.helpFold",
			"indianpoker.helpCheck",
			"indianpoker.helpCall",
			"indianpoker.helpBet",
			"indianpoker.helpRaise",
			"indianpoker.helpAllIn",
		},
		SettingKeys: []string{"indianpoker.helpAnte", "indianpoker.helpBettingLimit", "indianpoker.helpMetaAI"},
	}))
}
