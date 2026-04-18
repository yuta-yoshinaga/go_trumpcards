package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewOldMaidCui コンストラクタ
func NewOldMaidCui() *genericCuiGame {
	omc := controller.NewOldMaidCuiController(
		usecase.NewOldMaidInteractor(domain.NewDefaultOldMaid(), new(presenter.OldMaidCuiPresenter)),
	)
	return newCuiGame(omc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "oldmaid.helpTitle",
		CommandKeys: []string{"oldmaid.helpDraw", "oldmaid.helpShuffle", "oldmaid.helpReorder"},
		SettingKeys: []string{"oldmaid.helpSetMode", "oldmaid.helpSetPlacement", "oldmaid.helpSetMemoryAI"},
	}))
}
