package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewDaifugoCui コンストラクタ
func NewDaifugoCui() *genericCuiGame {
	dgc := controller.NewDaifugoCuiController(
		usecase.NewDaifugoInteractor(domain.NewDefaultDaifugo(), new(presenter.DaifugoCuiPresenter)),
	)
	return newCuiGame(dgc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "daifugo.helpTitle",
		CommandKeys: []string{"daifugo.helpPlay", "daifugo.helpSort"},
		SettingKeys: []string{"daifugo.helpSetDifficulty", "daifugo.helpSetJoker", "daifugo.helpSetRule"},
	}))
}
