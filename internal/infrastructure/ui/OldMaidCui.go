package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
		"Old Maid (ババ抜き)",
		"",
		"Game commands:",
		"  d                    draw a card",
		"  s                    shuffle hand",
		"  ro [i0 i1 ...]       reorder hand",
		"",
		"Settings:",
		"  sm [0-1]             set mode (0=Normal, 1=JijiNuki)",
		"  sps [0-1]            set CPU placement strategy (0=OFF, 1=ON)",
		"  sma [0-1]            set CPU memory AI (0=OFF, 1=ON)",
		"",
		"Session:",
		"  r                    reset game",
		"  q                    quit",
		"  help, ?              show this help",
	}
}

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	RunCuiLoop(cui.omc, cui.HelpLines())
}
