package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewPigsTailCui コンストラクタ
func NewPigsTailCui() *genericCuiGame {
	players := []*domain.PigsTailPlayer{
		domain.NewPigsTailPlayer(true),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
	}
	pigsTail := domain.NewPigsTail(domain.NewTrumpCards(0), players)
	ptc := controller.NewPigsTailCuiController(
		usecase.NewPigsTailInteractor(pigsTail, new(presenter.PigsTailCuiPresenter)),
	)
	return newCuiGame(ptc, []string{
		i18n.T("pigtail.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pigtail.helpAction"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
