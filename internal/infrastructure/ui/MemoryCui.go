package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewMemoryCui コンストラクタ
func NewMemoryCui() *genericCuiGame {
	config := domain.DefaultMemoryConfig()
	players := []*domain.MemoryPlayer{
		domain.NewMemoryPlayer(true),
		domain.NewMemoryPlayer(false),
		domain.NewMemoryPlayer(false),
		domain.NewMemoryPlayer(false),
	}
	memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
	mc := controller.NewMemoryCuiController(usecase.NewMemoryInteractor(memory, new(presenter.MemoryCuiPresenter)))
	return newCuiGame(mc, []string{
		i18n.T("memory.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("memory.helpFlip"),
		i18n.T("memory.helpNext"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("memory.helpSetDifficulty"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
