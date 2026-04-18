package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewWhistCui コンストラクタ
func NewWhistCui() *genericCuiGame {
	wc := controller.NewWhistCuiController(usecase.NewWhistInteractor(domain.NewDefaultWhist(), new(presenter.WhistCuiPresenter)))
	return newCuiGame(wc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "whist.helpTitle",
		CommandKeys:       []string{"whist.helpPlay", "whist.helpNext", "whist.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"whist.helpSetDifficulty", "whist.helpSetLimit"},
	}))
}
