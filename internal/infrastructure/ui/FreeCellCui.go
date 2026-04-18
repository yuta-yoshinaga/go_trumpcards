package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFreeCellCui コンストラクタ
func NewFreeCellCui() *genericCuiGame {
	fc := controller.NewFreeCellCuiController(usecase.NewFreeCellInteractor(domain.NewDefaultFreeCell(), new(presenter.FreeCellCuiPresenter)))
	return newCuiGame(fc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "freecell.helpTitle",
		CommandKeys: []string{
			"freecell.helpMove",
			"freecell.helpMoveTF",
			"freecell.helpMoveTT",
			"freecell.helpMoveTC",
			"freecell.helpMoveCT",
			"freecell.helpMoveCF",
			"freecell.helpGiveUp",
			"freecell.helpHint",
			"freecell.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
