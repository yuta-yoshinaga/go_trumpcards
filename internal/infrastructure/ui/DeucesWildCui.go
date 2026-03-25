package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DeucesWildCui Deuces Wild CUIクラス
type DeucesWildCui struct {
	vc *controller.VideoPokerCuiController
}

// NewDeucesWildCui コンストラクタ
func NewDeucesWildCui() *DeucesWildCui {
	return &DeucesWildCui{
		vc: controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
			domain.NewDeucesWildVideoPoker(),
			new(presenter.VideoPokerCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *DeucesWildCui) Controller() CuiExecer { return cui.vc }

// HelpLines returns the game's help lines.
func (cui *DeucesWildCui) HelpLines() []string {
	return []string{
		i18n.T("deuceswild.helpTitle"),
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
func (cui *DeucesWildCui) Exec() {
	RunCuiLoop(cui.vc, cui.HelpLines())
}
