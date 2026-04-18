package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewShortDeckCui コンストラクタ
func NewShortDeckCui() *genericCuiGame {
	cfg := domain.DefaultShortDeckConfig()
	sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
	sc := controller.NewShortDeckCuiController(usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckCuiPresenter)))
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "shortdeck.helpTitle",
		CommandKeys: []string{
			"shortdeck.helpFold",
			"shortdeck.helpCheck",
			"shortdeck.helpCall",
			"shortdeck.helpBet",
			"shortdeck.helpRaise",
			"shortdeck.helpAllIn",
		},
		ExtraCommandLines: []string{
			"  rb                   rebuy",
			"  sr                   skip rebuy",
			"  ad                   add-on",
			"  sa                   skip add-on",
		},
		SettingKeys: []string{"shortdeck.helpBettingLimit", "shortdeck.helpTournament"},
		ExtraSettingLines: []string{
			"  sb <amount>          small blind (>=1)",
			"  bb <amount>          big blind (>=2)",
			"  lh <hands>           blind level-up hands (>=1)",
			"  ts [4|6|9]           table size",
		},
	}))
}
