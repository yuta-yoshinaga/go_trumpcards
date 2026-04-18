package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewHoldemCui コンストラクタ
func NewHoldemCui() *genericCuiGame {
	hc := controller.NewHoldemCuiController(usecase.NewHoldemInteractor(domain.NewDefaultHoldem(), new(presenter.HoldemCuiPresenter)))
	return newCuiGame(hc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "holdem.helpTitle",
		CommandKeys: []string{
			"holdem.helpFold",
			"holdem.helpCheck",
			"holdem.helpCall",
			"holdem.helpBet",
			"holdem.helpRaise",
			"holdem.helpAllIn",
		},
		ExtraCommandLines: []string{
			"  rb                   rebuy",
			"  sr                   skip rebuy",
			"  ad                   add-on",
			"  sa                   skip add-on",
		},
		SettingKeys: []string{"holdem.helpBettingLimit", "holdem.helpTournament"},
		ExtraSettingLines: []string{
			"  sb <amount>          small blind (>=1)",
			"  bb <amount>          big blind (>=2)",
			"  lh <hands>           blind level-up hands (>=1)",
			"  ts [4|6|9]           table size",
		},
	}))
}
