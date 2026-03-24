package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EuchreCui ユーカーCUIクラス
type EuchreCui struct {
	ec *controller.EuchreCuiController
}

// NewEuchreCui コンストラクタ
func NewEuchreCui() *EuchreCui {
	config := domain.DefaultEuchreConfig()
	players := []*domain.EuchrePlayer{
		domain.NewEuchrePlayer(true, 0),
		domain.NewEuchrePlayer(false, 1),
		domain.NewEuchrePlayer(false, 0),
		domain.NewEuchrePlayer(false, 1),
	}
	euchre := domain.NewEuchre(domain.NewTrumpCardsEuchre(), players, config)
	return &EuchreCui{
		ec: controller.NewEuchreCuiController(usecase.NewEuchreInteractor(euchre, new(presenter.EuchreCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *EuchreCui) Controller() CuiExecer { return cui.ec }

// HelpLines returns the game's help lines.
func (cui *EuchreCui) HelpLines() []string {
	return []string{
		i18n.T("euchre.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("euchre.helpOrderUp"),
		i18n.T("euchre.helpOrderUpAlone"),
		i18n.T("euchre.helpPass"),
		i18n.T("euchre.helpCall"),
		i18n.T("euchre.helpCallAlone"),
		i18n.T("euchre.helpDiscard"),
		i18n.T("euchre.helpPlay"),
		i18n.T("euchre.helpNext"),
		i18n.T("euchre.helpNextRound"),
		"  l                    action log",
		"",
		i18n.T("settings"),
		i18n.T("euchre.helpSetDifficulty"),
		i18n.T("euchre.helpSetLimit"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *EuchreCui) Exec() {
	RunCuiLoop(cui.ec, cui.HelpLines())
}
