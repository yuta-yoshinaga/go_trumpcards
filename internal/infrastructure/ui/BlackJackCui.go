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
			new(presenter.BlackJackCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *BlackJackCui) Controller() CuiExecer { return cui.bjc }

// HelpLines returns the game's help lines.
func (cui *BlackJackCui) HelpLines() []string {
	return []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"b N [ppBet] [t3Bet] [handCount]・・・bet (e.g. b 100 0 0 2)",
		"h・・・hit",
		"s・・・stand",
		"d・・・doubledown",
		"sp・・・split",
		"i・・・insurance",
		"di・・・decline insurance",
		"scc N・・・set CPU player count (0-3)",
	}
}

// Exec ゲーム実行
func (cui *BlackJackCui) Exec() {
	RunCuiLoop(cui.bjc, cui.HelpLines())
}
