package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaccaratCui バカラCUIクラス
type BaccaratCui struct {
	bc *controller.BaccaratCuiController
}

// NewBaccaratCui コンストラクタ
func NewBaccaratCui() *BaccaratCui {
	return &BaccaratCui{
		bc: controller.NewBaccaratCuiController(usecase.NewBaccaratInteractor(
			domain.NewDefaultBaccarat(),
			new(presenter.BaccaratCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *BaccaratCui) Controller() CuiExecer { return cui.bc }

// HelpLines returns the game's help lines.
func (cui *BaccaratCui) HelpLines() []string {
	return []string{
		"Baccarat (バカラ)",
		"",
		"Game commands:",
		"  b <n> <t>            bet (n=amount, t: 0=Player, 1=Banker, 2=Tie)",
		"  log                  action log",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *BaccaratCui) Exec() {
	RunCuiLoop(cui.bc, cui.HelpLines())
}
