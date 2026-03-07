package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BlackJackCui ブラックジャックCUIクラス
type BlackJackCui struct {
	bjc *controller.BlackJackCuiController
}

// NewBlackJackCui コンストラクタ
func NewBlackJackCui() *BlackJackCui {
	return &BlackJackCui{
		bjc: controller.NewBlackJackCuiController(usecase.NewBlackJackInteractor(
			domain.NewDefaultBlackJack(),
			presenter.NewBlackJackCuiPresenter(),
		)),
	}
}

// Exec ゲーム実行
func (cui *BlackJackCui) Exec() {
	RunCuiLoop(cui.bjc, []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"b N・・・bet (e.g. b 100)",
		"h・・・hit",
		"s・・・stand",
		"d・・・doubledown",
		"sp・・・split",
		"i・・・insurance",
		"di・・・decline insurance",
		"scc N・・・set CPU player count (0-3)",
	})
}
