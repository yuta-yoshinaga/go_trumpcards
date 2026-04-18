package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewEuchreCui コンストラクタ
func NewEuchreCui() *genericCuiGame {
	config := domain.DefaultEuchreConfig()
	players := []*domain.EuchrePlayer{
		domain.NewEuchrePlayer(true, 0),
		domain.NewEuchrePlayer(false, 1),
		domain.NewEuchrePlayer(false, 0),
		domain.NewEuchrePlayer(false, 1),
	}
	euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
	ec := controller.NewEuchreCuiController(usecase.NewEuchreInteractor(euchre, new(presenter.EuchreCuiPresenter)))
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
