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
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"b N T・・・bet (e.g. b 100 0) T: 0=Player, 1=Banker, 2=Tie",
		"log・・・action log",
	}
}

// Exec ゲーム実行
func (cui *BaccaratCui) Exec() {
	RunCuiLoop(cui.bc, cui.HelpLines())
}
