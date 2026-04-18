package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewWarCui コンストラクタ
func NewWarCui() *genericCuiGame {
	wc := controller.NewWarCuiController(
		usecase.NewWarInteractor(domain.NewDefaultWar(), new(presenter.WarCuiPresenter)),
	)
	return newCuiGame(wc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "war.helpTitle",
		CommandKeys: []string{"war.helpStep"},
		SettingKeys: []string{"war.helpSetMax"},
	}))
}
