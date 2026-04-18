package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewHeartsCui コンストラクタ
func NewHeartsCui() *genericCuiGame {
	hc := controller.NewHeartsCuiController(usecase.NewHeartsInteractor(domain.NewDefaultHearts(), new(presenter.HeartsCuiPresenter)))
	return newCuiGame(hc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "hearts.helpTitle",
		CommandKeys:       []string{"hearts.helpPass", "hearts.helpPlay", "hearts.helpNext", "hearts.helpNextRound"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"hearts.helpSetDifficulty", "hearts.helpSetLimit"},
	}))
}
