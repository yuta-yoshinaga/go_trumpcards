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

// Exec ゲーム実行
func (cui *OldMaidCui) Exec() {
	RunCuiLoop(cui.omc, []string{
		"コマンドを入力してください。",
		"q・・・quit",
		"r・・・reset",
		"d・・・draw (カードを引く)",
		"s・・・shuffle (手札をシャッフル)",
		"ro [i0 i1 ...]・・・reorder (手札を並べ替え)",
		"sm [0-1]・・・set mode (0=Normal, 1=JijiNuki)",
		"sps [0-1]・・・set CPU placement strategy (0=OFF, 1=ON)",
		"sma [0-1]・・・set CPU memory AI (0=OFF, 1=ON)",
	})
}
