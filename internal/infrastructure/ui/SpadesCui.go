package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpadesCui コンストラクタ
func NewSpadesCui() *genericCuiGame {
	config := domain.DefaultSpadesConfig()
	players := []*domain.SpadesPlayer{
		domain.NewSpadesPlayer(true),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
		domain.NewSpadesPlayer(false),
	}
	spades := domain.NewSpades(domain.NewTrumpCards(0), players, config)
	sc := controller.NewSpadesCuiController(usecase.NewSpadesInteractor(spades, new(presenter.SpadesCuiPresenter)))
	return newCuiGame(sc, []string{
		i18n.T("spades.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("spades.helpBid"),
		i18n.T("spades.helpPlay"),
		i18n.T("spades.helpNext"),
		i18n.T("spades.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("spades.helpSetDifficulty"),
		i18n.T("spades.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
