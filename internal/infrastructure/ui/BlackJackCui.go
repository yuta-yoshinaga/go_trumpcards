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
		"BlackJack (ブラックジャック)",
		"",
		"Game commands:",
		"  b <n> [perfect-pairs] [triple-7s] [hands]  bet (e.g. b 100 0 0 2)",
		"  h                    hit",
		"  s                    stand",
		"  d                    double down",
		"  sp                   split",
		"  i                    insurance",
		"  di                   decline insurance",
		"",
		"Settings:",
		"  scc <n>              set CPU count (0-3)",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *BlackJackCui) Exec() {
	RunCuiLoop(cui.bjc, cui.HelpLines())
}
