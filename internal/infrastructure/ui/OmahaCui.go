package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewOmahaCui コンストラクタ
func NewOmahaCui() *genericCuiGame {
	oc := controller.NewOmahaCuiController(usecase.NewOmahaInteractor(domain.NewDefaultOmaha(), new(presenter.OmahaCuiPresenter)))
	return newCuiGame(oc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "omaha.helpTitle",
		CommandKeys: []string{
			"omaha.helpFold",
			"omaha.helpCheck",
			"omaha.helpCall",
			"omaha.helpBet",
			"omaha.helpRaise",
			"omaha.helpAllIn",
		},
		ExtraCommandLines: []string{
			"  rb                   rebuy",
			"  sr                   skip rebuy",
			"  ad                   add-on",
			"  sa                   skip add-on",
		},
		SettingKeys: []string{"omaha.helpBettingLimit", "omaha.helpTournament"},
		ExtraSettingLines: []string{
			"  sb <amount>          small blind (>=1)",
			"  bb <amount>          big blind (>=2)",
			"  lh <hands>           blind level-up hands (>=1)",
			"  ts [4|6|9]           table size",
		},
	}))
}
