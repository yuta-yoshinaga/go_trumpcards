package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPinochleCui コンストラクタ
func NewPinochleCui() *genericCuiGame {
	pc := controller.NewPinochleCuiController(usecase.NewPinochleInteractor(domain.NewDefaultPinochle(), new(presenter.PinochleCuiPresenter)))
	return newCuiGame(pc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "pinochle.helpTitle",
		CommandKeys: []string{
			"pinochle.helpBid",
			"pinochle.helpPass",
			"pinochle.helpTrump",
			"pinochle.helpMeld",
			"pinochle.helpPlay",
			"pinochle.helpNext",
			"pinochle.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"pinochle.helpSetDifficulty", "pinochle.helpSetLimit"},
	}))
}
