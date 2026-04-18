package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPinochleCui コンストラクタ
func NewPinochleCui() *genericCuiGame {
	config := domain.DefaultPinochleConfig()
	players := []*domain.PinochlePlayer{
		domain.NewPinochlePlayer(true, 0),
		domain.NewPinochlePlayer(false, 1),
		domain.NewPinochlePlayer(false, 0),
		domain.NewPinochlePlayer(false, 1),
	}
	pinochle := domain.NewPinochle(domain.NewTrumpCardsPinochle(), players, config)
	pc := controller.NewPinochleCuiController(usecase.NewPinochleInteractor(pinochle, new(presenter.PinochleCuiPresenter)))
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
