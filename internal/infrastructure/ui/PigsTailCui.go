package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPigsTailCui コンストラクタ
func NewPigsTailCui() *genericCuiGame {
	ptc := controller.NewPigsTailCuiController(
		usecase.NewPigsTailInteractor(domain.NewDefaultPigsTail(), new(presenter.PigsTailCuiPresenter)),
	)
	return newCuiGame(ptc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "pigtail.helpTitle",
		CommandKeys: []string{"pigtail.helpAction"},
	}))
}
