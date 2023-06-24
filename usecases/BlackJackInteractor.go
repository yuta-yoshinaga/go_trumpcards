package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
)

// BlackJackInteractorIF ブラックジャックインタラクターインタフェース
type BlackJackInteractorIF interface {
	Reset() string
	Hit() string
	Stand() string
}

// BlackJackInteractor ブラックジャックインタラクタークラス
type BlackJackInteractor struct {
	bj  *entities.BlackJack
	bjp presenters.BlackJackPresenter
}

// NewBlackJackInteractor コンストラクタ
func NewBlackJackInteractor(bjp presenters.BlackJackPresenter) *BlackJackInteractor {
	return &BlackJackInteractor{
		bj:  entities.NewBlackJack(entities.NewTrumpCards(0), entities.NewBlackJackPlayer(), entities.NewBlackJackPlayer()),
		bjp: bjp,
	}
}

// Reset ゲーム初期化
func (bi *BlackJackInteractor) Reset() string {
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj)
}

// Hit ヒット
func (bi *BlackJackInteractor) Hit() string {
	bi.bj.PlayerHit()
	return bi.bjp.Output(bi.bj)
}

// Stand スタンド
func (bi *BlackJackInteractor) Stand() string {
	bi.bj.PlayerStand()
	return bi.bjp.Output(bi.bj)
}
