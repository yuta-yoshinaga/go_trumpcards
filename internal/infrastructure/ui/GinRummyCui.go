package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewGinRummyCui コンストラクタ
func NewGinRummyCui() *genericCuiGame {
	config := domain.DefaultGinRummyConfig()
	players := []*domain.GinRummyPlayer{
		domain.NewGinRummyPlayer(true),
		domain.NewGinRummyPlayer(false),
	}
	gr := domain.NewGinRummy(domain.NewTrumpCards(0), players, config)
	cc := controller.NewGinRummyCuiController(usecase.NewGinRummyInteractor(gr, new(presenter.GinRummyCuiPresenter)))
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
