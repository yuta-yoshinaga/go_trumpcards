package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewTwoTenJackCui コンストラクタ
func NewTwoTenJackCui() *genericCuiGame {
	tc := controller.NewTwoTenJackCuiController(usecase.NewTwoTenJackInteractor(domain.NewDefaultTwoTenJack(), new(presenter.TwoTenJackCuiPresenter)))
	return newCuiGame(tc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "twotenjack.helpTitle",
		CommandKeys: []string{
			"twotenjack.helpDeclare",
			"twotenjack.helpPlay",
			"twotenjack.helpNext",
			"twotenjack.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"twotenjack.helpSetDifficulty", "twotenjack.helpSetLimit"},
	}))
}
