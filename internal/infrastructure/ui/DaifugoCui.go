package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewDaifugoCui コンストラクタ
func NewDaifugoCui() *genericCuiGame {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
	dgc := controller.NewDaifugoCuiController(
		usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoCuiPresenter)),
	)
	return newCuiGame(dgc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "daifugo.helpTitle",
		CommandKeys: []string{"daifugo.helpPlay", "daifugo.helpSort"},
		SettingKeys: []string{"daifugo.helpSetDifficulty", "daifugo.helpSetJoker", "daifugo.helpSetRule"},
	}))
}
