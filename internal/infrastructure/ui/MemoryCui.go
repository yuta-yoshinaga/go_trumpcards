package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MemoryCui 神経衰弱CUIクラス
type MemoryCui struct {
	mc *controller.MemoryCuiController
}

// NewMemoryCui コンストラクタ
func NewMemoryCui() *MemoryCui {
	config := domain.DefaultMemoryConfig()
	players := []*domain.MemoryPlayer{
		domain.NewMemoryPlayer(true),
		domain.NewMemoryPlayer(false),
		domain.NewMemoryPlayer(false),
		domain.NewMemoryPlayer(false),
	}
	memory := domain.NewMemory(domain.NewTrumpCards(0), players, config)
	return &MemoryCui{
		mc: controller.NewMemoryCuiController(usecase.NewMemoryInteractor(memory, new(presenter.MemoryCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *MemoryCui) Controller() CuiExecer { return cui.mc }

// HelpLines returns the game's help lines.
func (cui *MemoryCui) HelpLines() []string {
	return []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"f <pos>・・・flip card at position",
		"n・・・next (resolve flip)",
		"sd <0-2>・・・set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
		"l・・・action log",
	}
}

// Exec ゲーム実行
func (cui *MemoryCui) Exec() {
	RunCuiLoop(cui.mc, cui.HelpLines())
}
