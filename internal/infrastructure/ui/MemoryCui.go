package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewMemoryCui コンストラクタ
func NewMemoryCui() *genericCuiGame {
	mc := controller.NewMemoryCuiController(usecase.NewMemoryInteractor(domain.NewDefaultMemory(), new(presenter.MemoryCuiPresenter)))
	return newCuiGame(mc, BuildCuiHelp(CuiHelpSpec{
		TitleKey:          "memory.helpTitle",
		CommandKeys:       []string{"memory.helpFlip", "memory.helpNext"},
		ExtraCommandLines: []string{"  l                    action log"},
		SettingKeys:       []string{"memory.helpSetDifficulty"},
	}))
}
