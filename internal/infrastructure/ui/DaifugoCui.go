package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
		"Daifugo / Great Fool (大富豪)",
		"",
		"Game commands:",
		"  p [index ...]        play cards (no index = pass)",
		"  sort [0-2]           sort hand (0=strength, 1=suit, 2=number)",
		"",
		"Settings:",
		"  sd [0-2]             CPU difficulty (0=Normal, 1=Easy, 2=Hard)",
		"  sj [0-2]             joker count",
		"  sr <rule> <0|1>      toggle local rule",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *DaifugoCui) Exec() {
	RunCuiLoop(cui.dgc, cui.HelpLines())
}
