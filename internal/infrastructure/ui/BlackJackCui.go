package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *genericCuiGame {
	bjc := controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
		domain.NewDefaultBlackJack(),
		new(presenter.BlackJackCuiPresenter),
	))
	return newCuiGame(bjc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "blackjack.helpTitle",
		CommandKeys: []string{
			"blackjack.helpBet",
			"blackjack.helpHit",
			"blackjack.helpStand",
			"blackjack.helpDouble",
			"blackjack.helpSplit",
			"blackjack.helpInsurance",
			"blackjack.helpDeclineInsurance",
		},
		SettingKeys: []string{"blackjack.helpSetCpuCount"},
	}))
}
