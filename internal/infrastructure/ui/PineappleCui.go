package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPineappleCui コンストラクタ
func NewPineappleCui() *genericCuiGame {
	pc := controller.NewPineappleCuiController(usecase.NewPineappleInteractor(domain.NewDefaultPineapple(), new(presenter.PineappleCuiPresenter)))
	return newCuiGame(pc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "pineapple.helpTitle",
		CommandKeys: []string{
			"pineapple.helpFold",
			"pineapple.helpCheck",
			"pineapple.helpCall",
			"pineapple.helpBet",
			"pineapple.helpRaise",
			"pineapple.helpAllIn",
		},
		ExtraCommandLines: []string{
			"  d <index>            discard",
			"  rb                   rebuy",
			"  sr                   skip rebuy",
			"  ad                   add-on",
			"  sa                   skip add-on",
		},
		SettingKeys: []string{"pineapple.helpBettingLimit", "pineapple.helpTournament"},
		ExtraSettingLines: []string{
			"  sb <amount>          small blind (>=1)",
			"  bb <amount>          big blind (>=2)",
			"  lh <hands>           blind level-up hands (>=1)",
			"  ts [4|6|9]           table size",
		},
	}))
}
