package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewTwoTenJackCui コンストラクタ
func NewTwoTenJackCui() *genericCuiGame {
	config := domain.DefaultTwoTenJackConfig()
	players := []*domain.TwoTenJackPlayer{
		domain.NewTwoTenJackPlayer(true),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
		domain.NewTwoTenJackPlayer(false),
	}
	ttj := domain.NewTwoTenJack(domain.NewTrumpCards(0), players, config)
	tc := controller.NewTwoTenJackCuiController(usecase.NewTwoTenJackInteractor(ttj, new(presenter.TwoTenJackCuiPresenter)))
	return newCuiGame(tc, []string{
		i18n.T("twotenjack.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("twotenjack.helpDeclare"),
		i18n.T("twotenjack.helpPlay"),
		i18n.T("twotenjack.helpNext"),
		i18n.T("twotenjack.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("twotenjack.helpSetDifficulty"),
		i18n.T("twotenjack.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
