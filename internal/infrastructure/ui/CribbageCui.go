package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCribbageCui コンストラクタ
func NewCribbageCui() *genericCuiGame {
	config := domain.DefaultCribbageConfig()
	players := []*domain.CribbagePlayer{
		domain.NewCribbagePlayer(true),
		domain.NewCribbagePlayer(false),
	}
	g := domain.NewCribbage(domain.NewTrumpCards(0), players, config)
	cc := controller.NewCribbageCuiController(usecase.NewCribbageInteractor(g, new(presenter.CribbageCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "cribbage.helpTitle",
		CommandKeys:       []string{"cribbage.helpDiscard", "cribbage.helpPeg", "cribbage.helpGo", "cribbage.helpShowNext", "cribbage.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"cribbage.helpSetDifficulty", "cribbage.helpSetLimit"},
	}))
}
