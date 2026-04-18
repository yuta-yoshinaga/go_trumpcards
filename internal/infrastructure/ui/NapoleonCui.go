package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewNapoleonCui コンストラクタ
func NewNapoleonCui() *genericCuiGame {
	nc := controller.NewNapoleonCuiController(usecase.NewNapoleonInteractor(domain.NewDefaultNapoleon(), new(presenter.NapoleonCuiPresenter)))
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
