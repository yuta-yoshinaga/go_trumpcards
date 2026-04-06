package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewGoFishCui コンストラクタ
func NewGoFishCui() *genericCuiGame {
	players := []*domain.GoFishPlayer{
		domain.NewGoFishPlayer(true),
		domain.NewGoFishPlayer(false),
		domain.NewGoFishPlayer(false),
		domain.NewGoFishPlayer(false),
	}
	goFish := domain.NewGoFish(domain.NewTrumpCards(0), players)
	gfc := controller.NewGoFishCuiController(
		usecase.NewGoFishInteractor(goFish, new(presenter.GoFishCuiPresenter)),
	)
	return newCuiGame(gfc, []string{
		i18n.T("gofish.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("gofish.helpAsk"),
		"",
		i18n.T("settings"),
		i18n.T("gofish.helpSetDifficulty"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
