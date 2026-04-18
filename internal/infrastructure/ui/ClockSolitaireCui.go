package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewClockSolitaireCui コンストラクタ
func NewClockSolitaireCui() *genericCuiGame {
	cc := controller.NewClockSolitaireCuiController(usecase.NewClockSolitaireInteractor(domain.NewDefaultClockSolitaire(), new(presenter.ClockSolitaireCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "clocksolitaire.helpTitle",
		CommandKeys:       []string{"clocksolitaire.helpStep", "clocksolitaire.helpAutoPlay"},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
