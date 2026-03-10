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
			presenter.NewBaccaratCuiPresenter(),
		)),
	}
}

// Exec ゲーム実行
func (cui *BaccaratCui) Exec() {
	RunCuiLoop(cui.bc, []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"b N T・・・bet (e.g. b 100 0) T: 0=Player, 1=Banker, 2=Tie",
		"log・・・action log",
	})
}
