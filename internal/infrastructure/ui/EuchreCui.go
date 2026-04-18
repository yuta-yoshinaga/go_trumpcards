package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewEuchreCui コンストラクタ
func NewEuchreCui() *genericCuiGame {
	ec := controller.NewEuchreCuiController(usecase.NewEuchreInteractor(domain.NewDefaultEuchre(), new(presenter.EuchreCuiPresenter)))
	return newCuiGame(ec, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "euchre.helpTitle",
		CommandKeys: []string{
			"euchre.helpOrderUp",
			"euchre.helpOrderUpAlone",
			"euchre.helpPass",
			"euchre.helpCall",
			"euchre.helpCallAlone",
			"euchre.helpDiscard",
			"euchre.helpPlay",
			"euchre.helpNext",
			"euchre.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"euchre.helpSetDifficulty", "euchre.helpSetLimit"},
	}))
}
