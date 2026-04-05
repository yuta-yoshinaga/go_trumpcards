package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPinochleCui コンストラクタ
func NewPinochleCui() *genericCuiGame {
	config := domain.DefaultPinochleConfig()
	players := []*domain.PinochlePlayer{
		domain.NewPinochlePlayer(true, 0),
		domain.NewPinochlePlayer(false, 1),
		domain.NewPinochlePlayer(false, 0),
		domain.NewPinochlePlayer(false, 1),
	}
	pinochle := domain.NewPinochle(domain.NewTrumpCardsPinochle(), players, config)
	pc := controller.NewPinochleCuiController(usecase.NewPinochleInteractor(pinochle, new(presenter.PinochleCuiPresenter)))
	return newCuiGame(pc, []string{
		i18n.T("pinochle.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pinochle.helpBid"),
		i18n.T("pinochle.helpPass"),
		i18n.T("pinochle.helpTrump"),
		i18n.T("pinochle.helpMeld"),
		i18n.T("pinochle.helpPlay"),
		i18n.T("pinochle.helpNext"),
		i18n.T("pinochle.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("pinochle.helpSetDifficulty"),
		i18n.T("pinochle.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
