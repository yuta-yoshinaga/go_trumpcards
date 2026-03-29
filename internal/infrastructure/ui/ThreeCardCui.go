package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardCui スリーカードポーカーCUIクラス
type ThreeCardCui struct {
	tc *controller.ThreeCardCuiController
}

// NewThreeCardCui コンストラクタ
func NewThreeCardCui() *ThreeCardCui {
	return &ThreeCardCui{
		tc: controller.NewThreeCardCuiController(usecase.NewThreeCardInteractor(
			domain.NewDefaultThreeCard(),
			new(presenter.ThreeCardCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *ThreeCardCui) Controller() CuiExecer { return cui.tc }

// HelpLines returns the game's help lines.
func (cui *ThreeCardCui) HelpLines() []string {
	return []string{
		i18n.T("threecard.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("threecard.helpBet"),
		i18n.T("threecard.helpPlay"),
		i18n.T("threecard.helpFold"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *ThreeCardCui) Exec() {
	RunCuiLoop(cui.tc, cui.HelpLines())
}
