package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapoleonCui ナポレオンCUIクラス
type NapoleonCui struct {
	nc *controller.NapoleonCuiController
}

// NewNapoleonCui コンストラクタ
func NewNapoleonCui() *NapoleonCui {
	config := domain.DefaultNapoleonConfig()
	players := []*domain.NapoleonPlayer{
		domain.NewNapoleonPlayer(true),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
		domain.NewNapoleonPlayer(false),
	}
	napoleon := domain.NewNapoleon(domain.NewTrumpCards(1), players, config)
	return &NapoleonCui{
		nc: controller.NewNapoleonCuiController(usecase.NewNapoleonInteractor(napoleon, new(presenter.NapoleonCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *NapoleonCui) Controller() CuiExecer { return cui.nc }

// HelpLines returns the game's help lines.
func (cui *NapoleonCui) HelpLines() []string {
	return []string{
		i18n.T("napoleon.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("napoleon.helpBid"),
		i18n.T("napoleon.helpTrump"),
		i18n.T("napoleon.helpExchange"),
		i18n.T("napoleon.helpPlay"),
		i18n.T("napoleon.helpNext"),
		i18n.T("napoleon.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("napoleon.helpSetDifficulty"),
		i18n.T("napoleon.helpSetLimit"),
		i18n.T("napoleon.helpSetMinBid"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *NapoleonCui) Exec() {
	RunCuiLoop(cui.nc, cui.HelpLines())
}
