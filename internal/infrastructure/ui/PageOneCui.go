package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPageOneCui コンストラクタ
func NewPageOneCui() *genericCuiGame {
	cc := controller.NewPageOneCuiController(usecase.NewPageOneInteractor(domain.NewDefaultPageOne(), new(presenter.PageOneCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "pageone.helpTitle",
		CommandKeys: []string{
			"pageone.helpPlay",
			"pageone.helpDraw",
			"pageone.helpDeclare",
			"pageone.helpSkip",
			"pageone.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"pageone.helpSetDifficulty", "pageone.helpSetLimit"},
	}))
}
