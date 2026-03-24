package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PyramidCui ピラミッドCUIクラス
type PyramidCui struct {
	pc *controller.PyramidCuiController
}

// NewPyramidCui コンストラクタ
func NewPyramidCui() *PyramidCui {
	pyramid := domain.NewPyramid(domain.NewTrumpCards(0))
	return &PyramidCui{
		pc: controller.NewPyramidCuiController(usecase.NewPyramidInteractor(pyramid, new(presenter.PyramidCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *PyramidCui) Controller() CuiExecer { return cui.pc }

// HelpLines returns the game's help lines.
func (cui *PyramidCui) HelpLines() []string {
	return []string{
		i18n.T("pyramid.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pyramid.helpDraw"),
		i18n.T("pyramid.helpRemoveKing"),
		i18n.T("pyramid.helpRemovePair"),
		i18n.T("pyramid.helpRemoveWaste"),
		i18n.T("pyramid.helpRemoveWasteKing"),
		i18n.T("pyramid.helpGiveUp"),
		i18n.T("pyramid.helpHint"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *PyramidCui) Exec() {
	RunCuiLoop(cui.pc, cui.HelpLines())
}
