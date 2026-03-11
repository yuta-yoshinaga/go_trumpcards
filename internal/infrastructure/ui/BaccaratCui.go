package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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
		i18n.T("baccarat.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("baccarat.helpBet"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *BaccaratCui) Exec() {
	RunCuiLoop(cui.bc, cui.HelpLines())
}
