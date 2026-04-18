package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewGinRummyCui コンストラクタ
func NewGinRummyCui() *genericCuiGame {
	cc := controller.NewGinRummyCuiController(usecase.NewGinRummyInteractor(domain.NewDefaultGinRummy(), new(presenter.GinRummyCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "ginrummy.helpTitle",
		CommandKeys: []string{
			"ginrummy.helpDrawStock",
			"ginrummy.helpDrawDiscard",
			"ginrummy.helpDiscard",
			"ginrummy.helpKnock",
			"ginrummy.helpLayoff",
			"ginrummy.helpNextRound",
		},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"ginrummy.helpSetDifficulty", "ginrummy.helpSetLimit"},
	}))
}
