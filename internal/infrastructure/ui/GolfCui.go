package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewGolfCui コンストラクタ
func NewGolfCui() *genericCuiGame {
	golf := domain.NewGolf(domain.NewTrumpCards(0))
	gc := controller.NewGolfCuiController(usecase.NewGolfInteractor(golf, new(presenter.GolfCuiPresenter)))
	return newCuiGame(gc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "golf.helpTitle",
		CommandKeys:       []string{"golf.helpDraw", "golf.helpRemove", "golf.helpGiveUp", "golf.helpHint"},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
