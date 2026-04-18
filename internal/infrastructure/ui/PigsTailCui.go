package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPigsTailCui コンストラクタ
func NewPigsTailCui() *genericCuiGame {
	players := []*domain.PigsTailPlayer{
		domain.NewPigsTailPlayer(true),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
	}
	pigsTail := domain.NewPigsTail(domain.NewTrumpCards(0), players)
	ptc := controller.NewPigsTailCuiController(
		usecase.NewPigsTailInteractor(pigsTail, new(presenter.PigsTailCuiPresenter)),
	)
	return newCuiGame(ptc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "pigtail.helpTitle",
		CommandKeys: []string{"pigtail.helpAction"},
	}))
}
