package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewRazzCui コンストラクタ
func NewRazzCui() *genericCuiGame {
	sc := controller.NewSevenCardStudCuiController(usecase.NewSevenCardStudInteractor(domain.NewDefaultRazz(), new(presenter.SevenCardStudCuiPresenter)))
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "razz.helpTitle",
		CommandKeys: []string{
			"sevencardstud.helpFold",
			"sevencardstud.helpCheck",
			"sevencardstud.helpCall",
			"sevencardstud.helpBet",
			"sevencardstud.helpRaise",
			"sevencardstud.helpAllIn",
		},
		ExtraCommandLines: []string{
			"  rb                   rebuy",
			"  sr                   skip rebuy",
			"  ad                   add-on",
			"  sa                   skip add-on",
		},
		SettingKeys: []string{"sevencardstud.helpBettingLimit", "sevencardstud.helpTournament"},
		ExtraSettingLines: []string{
			"  ante <amount>        ante (>=1)",
			"  bi <amount>          bring-in (>=1)",
			"  sb <amount>          small bet (>=1)",
			"  bb <amount>          big bet (>=1)",
			"  lh <hands>           ante level-up hands (>=1)",
			"  ts [2-7]             table size",
		},
	}))
}
