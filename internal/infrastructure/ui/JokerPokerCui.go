package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// JokerPokerCui Joker Poker CUIクラス
type JokerPokerCui struct {
	vc *controller.VideoPokerCuiController
}

// NewJokerPokerCui コンストラクタ
func NewJokerPokerCui() *JokerPokerCui {
	return &JokerPokerCui{
		vc: controller.NewVideoPokerCuiController(usecase.NewVideoPokerInteractor(
			domain.NewJokerPokerVideoPoker(),
			new(presenter.VideoPokerCuiPresenter),
		)),
	}
}

// Controller returns the game controller.
func (cui *JokerPokerCui) Controller() CuiExecer { return cui.vc }

// HelpLines returns the game's help lines.
func (cui *JokerPokerCui) HelpLines() []string {
	return []string{
		i18n.T("jokerpoker.helpTitle"),
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
func (cui *JokerPokerCui) Exec() {
	RunCuiLoop(cui.vc, cui.HelpLines())
}
