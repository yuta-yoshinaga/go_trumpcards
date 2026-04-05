package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSevensCui コンストラクタ
func NewSevensCui() *genericCuiGame {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
	sgc := controller.NewSevensCuiController(
		usecase.NewSevensInteractor(sevens, new(presenter.SevensCuiPresenter)),
	)
	return newCuiGame(sgc, []string{
		i18n.T("sevens.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("sevens.helpPlay"),
		"",
		i18n.T("session"),
		"  r [tunnel] [joker=N] [strategy] [passes=N]  reset with options",
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
