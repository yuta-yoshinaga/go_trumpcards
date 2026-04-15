package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPageOneCui コンストラクタ
func NewPageOneCui() *genericCuiGame {
	config := domain.DefaultPageOneConfig()
	players := []*domain.PageOnePlayer{
		domain.NewPageOnePlayer(true),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
		domain.NewPageOnePlayer(false),
	}
	po := domain.NewPageOne(domain.NewTrumpCards(0), players, config)
	cc := controller.NewPageOneCuiController(usecase.NewPageOneInteractor(po, new(presenter.PageOneCuiPresenter)))
	return newCuiGame(cc, []string{
		i18n.T("pageone.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pageone.helpPlay"),
		i18n.T("pageone.helpDraw"),
		i18n.T("pageone.helpDeclare"),
		i18n.T("pageone.helpSkip"),
		i18n.T("pageone.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("pageone.helpSetDifficulty"),
		i18n.T("pageone.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
