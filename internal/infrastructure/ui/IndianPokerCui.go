package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IndianPokerCui インディアンポーカーCUIクラス
type IndianPokerCui struct {
	ipc *controller.IndianPokerCuiController
}

// NewIndianPokerCui コンストラクタ
func NewIndianPokerCui() *IndianPokerCui {
	cfg := domain.DefaultIndianPokerConfig()
	ip := domain.NewIndianPoker(domain.NewTrumpCards(0), domain.NewIndianPokerPlayers(), cfg)
	return &IndianPokerCui{
		ipc: controller.NewIndianPokerCuiController(usecase.NewIndianPokerInteractor(ip, new(presenter.IndianPokerCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *IndianPokerCui) Controller() CuiExecer { return cui.ipc }

// HelpLines returns the game's help lines.
func (cui *IndianPokerCui) HelpLines() []string {
	return []string{
		i18n.T("indianpoker.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("indianpoker.helpFold"),
		i18n.T("indianpoker.helpCheck"),
		i18n.T("indianpoker.helpCall"),
		i18n.T("indianpoker.helpBet"),
		i18n.T("indianpoker.helpRaise"),
		i18n.T("indianpoker.helpAllIn"),
		"",
		i18n.T("settings"),
		i18n.T("indianpoker.helpAnte"),
		i18n.T("indianpoker.helpBettingLimit"),
		i18n.T("indianpoker.helpMetaAI"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *IndianPokerCui) Exec() {
	RunCuiLoop(cui.ipc, cui.HelpLines())
}
