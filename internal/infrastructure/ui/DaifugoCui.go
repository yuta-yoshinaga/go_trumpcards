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
			usecase.NewDaifugoInteractor(daifugo, presenter.NewDaifugoCuiPresenter()),
		),
	}
}

// Exec ゲーム実行
func (cui *DaifugoCui) Exec() {
	RunCuiLoop(cui.dgc, []string{
		"コマンドを入力してください。",
		"q・・・quit",
		"r・・・reset",
		"p [インデックス...]・・・カードを出す (インデックスなしでパス)",
	})
}
