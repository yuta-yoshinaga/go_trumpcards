package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewJokerPokerCui コンストラクタ
func NewJokerPokerCui() *genericCuiGame {
	vc := controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
		domain.NewJokerPokerVideoPoker(),
		new(presenter.VideoPokerCuiPresenter),
	))
	return newCuiGame(vc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "jokerpoker.helpTitle",
		CommandKeys:       []string{"videopoker.helpBet", "videopoker.helpHold"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
