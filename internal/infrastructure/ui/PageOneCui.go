package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPageOneCui コンストラクタ
func NewPageOneCui() *genericCuiGame {
	config := domain.DefaultPageOneConfig()
	players := []*domain.PageOnePlayer{
		domain.NewPageOnePlayer(true),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
	}
	po := domain.NewPageOne(domain.NewTrumpCards(0), players, config)
	cc := controller.NewPageOneCuiController(usecase.NewPageOneInteractor(po, new(presenter.PageOneCuiPresenter)))
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
