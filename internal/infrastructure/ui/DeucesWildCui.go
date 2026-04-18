package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewDeucesWildCui コンストラクタ
func NewDeucesWildCui() *genericCuiGame {
	vc := controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
		domain.NewDeucesWildVideoPoker(),
		new(presenter.VideoPokerCuiPresenter),
	))
	return newCuiGame(vc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "deuceswild.helpTitle",
		CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
