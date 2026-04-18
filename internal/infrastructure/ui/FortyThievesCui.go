package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFortyThievesCui コンストラクタ
func NewFortyThievesCui() *genericCuiGame {
	fc := controller.NewFortyThievesCuiController(usecase.NewFortyThievesInteractor(domain.NewDefaultFortyThieves(), new(presenter.FortyThievesCuiPresenter)))
	return newCuiGame(fc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "fortythieves.helpTitle",
		CommandKeys: []string{
			"fortythieves.helpDraw",
			"fortythieves.helpMove",
			"fortythieves.helpMoveWF",
			"fortythieves.helpMoveTF",
			"fortythieves.helpMoveTT",
			"fortythieves.helpGiveUp",
			"fortythieves.helpHint",
			"fortythieves.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
