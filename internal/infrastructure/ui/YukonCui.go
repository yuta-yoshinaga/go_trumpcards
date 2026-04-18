package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewYukonCui コンストラクタ
func NewYukonCui() *genericCuiGame {
	yc := controller.NewYukonCuiController(usecase.NewYukonInteractor(domain.NewDefaultYukon(), new(presenter.YukonCuiPresenter)))
	return newCuiGame(yc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "yukon.helpTitle",
		CommandKeys: []string{
			"yukon.helpMove",
			"yukon.helpMoveTF",
			"yukon.helpMoveTT",
			"yukon.helpGiveUp",
			"yukon.helpHint",
			"yukon.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
