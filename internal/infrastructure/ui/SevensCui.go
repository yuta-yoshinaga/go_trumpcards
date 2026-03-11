package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevensCui 7並べCUIクラス
type SevensCui struct {
	sgc *controller.SevensCuiController
}

// NewSevensCui コンストラクタ
func NewSevensCui() *SevensCui {
	config := domain.DefaultSevensConfig()
	players := []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
	sevens := domain.NewSevens(domain.NewTrumpCards(config.JokerCount), players, config)
	return &SevensCui{
		sgc: controller.NewSevensCuiController(
			usecase.NewSevensInteractor(sevens, new(presenter.SevensCuiPresenter)),
		),
	}
}

// Exec ゲーム実行
func (cui *SevensCui) Exec() {
	RunCuiLoop(cui.sgc, []string{
		"コマンドを入力してください。",
		"q・・・quit",
		"r [tunnel] [joker=N] [strategy] [passes=N]・・・reset (オプションルール設定)",
		"p [インデックス]・・・カードを出す (インデックスなしでパス)",
	})
}
