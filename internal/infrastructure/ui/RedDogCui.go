package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewRedDogCui コンストラクタ
func NewRedDogCui() *genericCuiGame {
	rc := controller.NewRedDogCuiController(usecase.NewRedDogInteractor(
		domain.NewDefaultRedDog(),
		new(presenter.RedDogCuiPresenter),
	))
	return newCuiGame(rc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "reddog.helpTitle",
		CommandKeys:       []string{"reddog.helpBet", "reddog.helpRaise", "reddog.helpStay"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
