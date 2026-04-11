package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewWarCui コンストラクタ
func NewWarCui() *genericCuiGame {
	players := []*domain.WarPlayer{
		domain.NewWarPlayer(true),
		domain.NewWarPlayer(false),
	}
	config := domain.DefaultWarConfig()
	game := domain.NewWar(domain.NewTrumpCards(0), players, config)
	wc := controller.NewWarCuiController(
		usecase.NewWarInteractor(game, new(presenter.WarCuiPresenter)),
	)
	return newCuiGame(wc, []string{
		i18n.T("war.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("war.helpStep"),
		"",
		i18n.T("settings"),
		i18n.T("war.helpSetMax"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
