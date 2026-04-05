package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewHeartsCui コンストラクタ
func NewHeartsCui() *genericCuiGame {
	config := domain.DefaultHeartsConfig()
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
	hc := controller.NewHeartsCuiController(usecase.NewHeartsInteractor(hearts, new(presenter.HeartsCuiPresenter)))
	return newCuiGame(hc, []string{
		i18n.T("hearts.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("hearts.helpPass"),
		i18n.T("hearts.helpPlay"),
		i18n.T("hearts.helpNext"),
		i18n.T("hearts.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("hearts.helpSetDifficulty"),
		i18n.T("hearts.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
