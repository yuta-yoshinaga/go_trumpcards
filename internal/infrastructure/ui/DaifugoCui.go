package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DaifugoCui 大富豪CUIクラス
type DaifugoCui struct {
	dgc *controller.DaifugoCuiController
}

// NewDaifugoCui コンストラクタ
func NewDaifugoCui() *DaifugoCui {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	daifugo := domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
	return &DaifugoCui{
		dgc: controller.NewDaifugoCuiController(
			usecase.NewDaifugoInteractor(daifugo, new(presenter.DaifugoCuiPresenter)),
		),
	}
}

// Controller returns the game controller.
func (cui *DaifugoCui) Controller() CuiExecer { return cui.dgc }

// HelpLines returns the game's help lines.
func (cui *DaifugoCui) HelpLines() []string {
	return []string{
		i18n.T("daifugo.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("daifugo.helpPlay"),
		i18n.T("daifugo.helpSort"),
		"",
		i18n.T("settings"),
		i18n.T("daifugo.helpSetDifficulty"),
		i18n.T("daifugo.helpSetJoker"),
		i18n.T("daifugo.helpSetRule"),
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *DaifugoCui) Exec() {
	RunCuiLoop(cui.dgc, cui.HelpLines())
}
