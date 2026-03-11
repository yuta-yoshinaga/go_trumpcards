package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OldMaidCui ババ抜きCUIクラス
type OldMaidCui struct {
	omc *controller.OldMaidCuiController
}

// NewOldMaidCui コンストラクタ
func NewOldMaidCui() *OldMaidCui {
	players := []*domain.OldMaidPlayer{
		domain.NewOldMaidPlayer(true),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
		domain.NewOldMaidPlayer(false),
	}
	oldMaid := domain.NewOldMaid(domain.NewTrumpCards(1), players)
	return &OldMaidCui{
		omc: controller.NewOldMaidCuiController(
			usecase.NewOldMaidInteractor(oldMaid, new(presenter.OldMaidCuiPresenter)),
		),
	}
}

// Controller returns the game controller.
func (cui *OldMaidCui) Controller() CuiExecer { return cui.omc }

// HelpLines returns the game's help lines.
func (cui *OldMaidCui) HelpLines() []string {
	return []string{
		i18n.T("oldmaid.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("oldmaid.helpDraw"),
		i18n.T("oldmaid.helpShuffle"),
		i18n.T("oldmaid.helpReorder"),
		"",
		i18n.T("settings"),
		i18n.T("oldmaid.helpSetMode"),
		i18n.T("oldmaid.helpSetPlacement"),
		i18n.T("oldmaid.helpSetMemoryAI"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	RunCuiLoop(cui.omc, cui.HelpLines())
}
