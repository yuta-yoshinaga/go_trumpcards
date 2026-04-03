package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GolfCui ゴルフソリティアCUIクラス
type GolfCui struct {
	gc *controller.GolfCuiController
}

// NewGolfCui コンストラクタ
func NewGolfCui() *GolfCui {
	golf := domain.NewGolf(domain.NewTrumpCards(0))
	return &GolfCui{
		gc: controller.NewGolfCuiController(usecase.NewGolfInteractor(golf, new(presenter.GolfCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *GolfCui) Controller() CuiExecer { return cui.gc }

// HelpLines returns the game's help lines.
func (cui *GolfCui) HelpLines() []string {
	return []string{
		i18n.T("golf.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("golf.helpDraw"),
		i18n.T("golf.helpRemove"),
		i18n.T("golf.helpGiveUp"),
		i18n.T("golf.helpHint"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *GolfCui) Exec() {
	RunCuiLoop(cui.gc, cui.HelpLines())
}
