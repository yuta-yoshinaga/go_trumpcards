package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewNapoleonCui コンストラクタ
func NewNapoleonCui() *genericCuiGame {
	config := domain.DefaultNapoleonConfig()
	players := []*domain.NapoleonPlayer{
		domain.NewNapoleonPlayer(true),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
	}
	napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
	nc := controller.NewNapoleonCuiController(usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonCuiPresenter)))
	return newCuiGame(nc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "napoleon.helpTitle",
		CommandKeys: []string{
			"napoleon.helpBid",
			"napoleon.helpTrump",
			"napoleon.helpExchange",
			"napoleon.helpPlay",
			"napoleon.helpNext",
			"napoleon.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"napoleon.helpSetDifficulty", "napoleon.helpSetLimit", "napoleon.helpSetMinBid"},
	}))
}
