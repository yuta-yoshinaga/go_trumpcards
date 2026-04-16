package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewRazzCui コンストラクタ
func NewRazzCui() *genericCuiGame {
	cfg := domain.DefaultRazzConfig()
	r := domain.NewRazz(domain.NewTrumpCards(0), domain.NewSevenCardStudPlayersForTable(cfg.TableSize), cfg)
	sc := controller.NewSevenCardStudCuiController(usecase.NewSevenCardStudInteractor(r, new(presenter.SevenCardStudCuiPresenter)))
	return newCuiGame(sc, []string{
		i18n.T("razz.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("sevencardstud.helpFold"),
		i18n.T("sevencardstud.helpCheck"),
		i18n.T("sevencardstud.helpCall"),
		i18n.T("sevencardstud.helpBet"),
		i18n.T("sevencardstud.helpRaise"),
		i18n.T("sevencardstud.helpAllIn"),
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		i18n.T("settings"),
		i18n.T("sevencardstud.helpBettingLimit"),
		i18n.T("sevencardstud.helpTournament"),
		"  ante <amount>        ante (>=1)",
		"  bi <amount>          bring-in (>=1)",
		"  sb <amount>          small bet (>=1)",
		"  bb <amount>          big bet (>=1)",
		"  lh <hands>           ante level-up hands (>=1)",
		"  ts [2-7]             table size",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
