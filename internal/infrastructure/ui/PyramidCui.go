package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPyramidCui コンストラクタ
func NewPyramidCui() *genericCuiGame {
	pc := controller.NewPyramidCuiController(usecase.NewPyramidInteractor(domain.NewDefaultPyramid(), new(presenter.PyramidCuiPresenter)))
	return newCuiGame(pc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "pyramid.helpTitle",
		CommandKeys: []string{
			"pyramid.helpDraw",
			"pyramid.helpRemoveKing",
			"pyramid.helpRemovePair",
			"pyramid.helpRemoveWaste",
			"pyramid.helpRemoveWasteKing",
			"pyramid.helpGiveUp",
			"pyramid.helpHint",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
