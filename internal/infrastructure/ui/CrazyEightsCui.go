package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

type CrazyEightsCui struct {
	cc *controller.CrazyEightsCuiController
}

func NewCrazyEightsCui() *CrazyEightsCui {
	config := domain.DefaultCrazyEightsConfig()
	players := []*domain.CrazyEightsPlayer{
		domain.NewCrazyEightsPlayer(true),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
		domain.NewCrazyEightsPlayer(false),
	}
	ce := domain.NewCrazyEights(domain.NewTrumpCards(0), players, config)
	return &CrazyEightsCui{
		cc: controller.NewCrazyEightsCuiController(usecase.NewCrazyEightsInteractor(ce, new(presenter.CrazyEightsCuiPresenter))),
	}
}

func (cui *CrazyEightsCui) Controller() CuiExecer { return cui.cc }

func (cui *CrazyEightsCui) HelpLines() []string {
	return []string{
		i18n.T("crazyeights.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("crazyeights.helpPlay"),
		i18n.T("crazyeights.helpDraw"),
		i18n.T("crazyeights.helpSuit"),
		i18n.T("crazyeights.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("crazyeights.helpSetDifficulty"),
		i18n.T("crazyeights.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

func (cui *CrazyEightsCui) Exec() {
	RunCuiLoop(cui.cc, cui.HelpLines())
}
