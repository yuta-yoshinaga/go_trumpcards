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
	Bet(amount int) string
	DoubleDown() string
	Split() string
	Insurance() string
	DeclineInsurance() string
}

// BlackJackInteractor ブラックジャックインタラクタークラス
type BlackJackInteractor struct {
	bj  *entities.BlackJack
	bjp presenters.BlackJackPresenter
}

// NewBlackJackInteractor コンストラクタ
func NewBlackJackInteractor(bj *entities.BlackJack, bjp presenters.BlackJackPresenter) *BlackJackInteractor {
	return &BlackJackInteractor{
		bj:  bj,
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

// Bet ベット
func (bi *BlackJackInteractor) Bet(amount int) string {
	bi.bj.PlayerBet(amount)
	return bi.bjp.Output(bi.bj)
}

// DoubleDown ダブルダウン
func (bi *BlackJackInteractor) DoubleDown() string {
	bi.bj.PlayerDoubleDown()
	return bi.bjp.Output(bi.bj)
}

// Split スプリット
func (bi *BlackJackInteractor) Split() string {
	bi.bj.PlayerSplit()
	return bi.bjp.Output(bi.bj)
}

// Insurance インシュランス
func (bi *BlackJackInteractor) Insurance() string {
	bi.bj.PlayerInsurance()
	return bi.bjp.Output(bi.bj)
}

// DeclineInsurance インシュランス辞退
func (bi *BlackJackInteractor) DeclineInsurance() string {
	bi.bj.PlayerDeclineInsurance()
	return bi.bjp.Output(bi.bj)
}
