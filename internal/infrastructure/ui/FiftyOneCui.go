package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFiftyOneCui コンストラクタ
func NewFiftyOneCui() *genericCuiGame {
	players := []*domain.FiftyOnePlayer{
		domain.NewFiftyOnePlayer(true),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
		domain.NewFiftyOnePlayer(false),
	}
	fo := domain.NewFiftyOne(domain.NewTrumpCards(0), players)
	foc := controller.NewFiftyOneCuiController(
		usecase.NewFiftyOneInteractor(fo, new(presenter.FiftyOneCuiPresenter)),
	)
	return newCuiGame(foc, []string{
		i18n.T("fiftyone.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("fiftyone.helpPlay"),
		i18n.T("fiftyone.helpAll"),
		i18n.T("fiftyone.helpStop"),
		"",
		i18n.T("settings"),
		i18n.T("fiftyone.helpSetDifficulty"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
