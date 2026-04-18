package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPokerCui コンストラクタ
func NewPokerCui() *genericCuiGame {
	pc := controller.NewPokerCuiController(usecase.NewPokerInteractor(domain.NewDefaultPoker(), new(presenter.PokerCuiPresenter)))
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
