package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewWhistCui コンストラクタ
func NewWhistCui() *genericCuiGame {
	config := domain.DefaultWhistConfig()
	players := []*domain.WhistPlayer{
		domain.NewWhistPlayer(true, 0),
		domain.NewWhistPlayer(false, 1),
		domain.NewWhistPlayer(false, 0),
		domain.NewWhistPlayer(false, 1),
	}
	whist := domain.NewWhist(domain.NewTrumpCards(0), players, config)
	wc := controller.NewWhistCuiController(usecase.NewWhistInteractor(whist, new(presenter.WhistCuiPresenter)))
	return newCuiGame(wc, []string{
		i18n.T("whist.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("whist.helpPlay"),
		i18n.T("whist.helpNext"),
		i18n.T("whist.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("whist.helpSetDifficulty"),
		i18n.T("whist.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
