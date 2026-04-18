package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCaribbeanStudCui コンストラクタ
func NewCaribbeanStudCui() *genericCuiGame {
	cs := controller.NewCaribbeanStudCuiController(usecase.NewCaribbeanStudInteractor(
		domain.NewDefaultCaribbeanStud(),
		new(presenter.CaribbeanStudCuiPresenter),
	))
	return newCuiGame(cs, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "caribbeanstud.helpTitle",
		CommandKeys:       []string{"caribbeanstud.helpBet", "caribbeanstud.helpPlay", "caribbeanstud.helpFold"},
		ExtraCommandLines: []string{"  log                  action log"},
	}))
}
