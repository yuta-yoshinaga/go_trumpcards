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
			usecase.NewOldMaidInteractor(oldMaid, presenter.NewOldMaidCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	RunCuiLoop(cui.omc, []string{
		"コマンドを入力してください。",
		"q・・・quit",
		"r・・・reset",
		"d・・・draw (カードを引く)",
	})
}
