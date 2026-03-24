package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// VideoPokerCui ビデオポーカーCUIクラス
type VideoPokerCui struct {
	vc *controller.VideoPokerCuiController
}

// NewVideoPokerCui コンストラクタ
func NewVideoPokerCui() *VideoPokerCui {
	return &VideoPokerCui{
		vc: controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
			domain.NewDefaultVideoPoker(),
			new(presenter.VideoPokerCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *VideoPokerCui) Controller() CuiExecer { return cui.vc }

// HelpLines returns the game's help lines.
func (cui *VideoPokerCui) HelpLines() []string {
	return []string{
		i18n.T("videopoker.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("videopoker.helpBet"),
		i18n.T("videopoker.helpHold"),
		"  log                  action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *VideoPokerCui) Exec() {
	RunCuiLoop(cui.vc, cui.HelpLines())
}
