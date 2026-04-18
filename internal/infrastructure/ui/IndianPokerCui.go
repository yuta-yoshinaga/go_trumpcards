package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewIndianPokerCui コンストラクタ
func NewIndianPokerCui() *genericCuiGame {
	ipc := controller.NewIndianPokerCuiController(usecase.NewIndianPokerInteractor(domain.NewDefaultIndianPoker(), new(presenter.IndianPokerCuiPresenter)))
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
