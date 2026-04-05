package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewOhHellCui コンストラクタ
func NewOhHellCui() *genericCuiGame {
	config := domain.DefaultOhHellConfig()
	players := []*domain.OhHellPlayer{
		domain.NewOhHellPlayer(true),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
		domain.NewOhHellPlayer(false),
	}
	ohHell := domain.NewOhHell(domain.NewTrumpCards(0), players, config)
	oc := controller.NewOhHellCuiController(usecase.NewOhHellInteractor(ohHell, new(presenter.OhHellCuiPresenter)))
	return newCuiGame(oc, []string{
		i18n.T("ohhell.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("ohhell.helpBid"),
		i18n.T("ohhell.helpPlay"),
		i18n.T("ohhell.helpNext"),
		i18n.T("ohhell.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("ohhell.helpSetDifficulty"),
		i18n.T("ohhell.helpSetMaxHand"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
