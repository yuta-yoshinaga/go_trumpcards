package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
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
	bj  interfaces.BlackJackGame
	bjp presenter.BlackJackPresenter
}

// NewBlackJackInteractor コンストラクタ
func NewBlackJackInteractor(bj interfaces.BlackJackGame, bjp presenter.BlackJackPresenter) *BlackJackInteractor {
	if bj == nil {
		panic("BlackJackInteractor: bj must not be nil")
	}
	if bjp == nil {
		panic("BlackJackInteractor: bjp must not be nil")
	}
	return &BlackJackInteractor{
		bj:  bj,
		bjp: bjp,
	}
}

// Reset ゲーム初期化
func (bi *BlackJackInteractor) Reset() string {
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj, nil)
}

// Hit ヒット
func (bi *BlackJackInteractor) Hit() string {
	err := bi.bj.PlayerHit()
	return bi.bjp.Output(bi.bj, err)
}

// Stand スタンド
func (bi *BlackJackInteractor) Stand() string {
	err := bi.bj.PlayerStand()
	return bi.bjp.Output(bi.bj, err)
}

// Bet ベット
func (bi *BlackJackInteractor) Bet(amount int) string {
	err := bi.bj.PlayerBet(amount)
	return bi.bjp.Output(bi.bj, err)
}

// DoubleDown ダブルダウン
func (bi *BlackJackInteractor) DoubleDown() string {
	err := bi.bj.PlayerDoubleDown()
	return bi.bjp.Output(bi.bj, err)
}

// Split スプリット
func (bi *BlackJackInteractor) Split() string {
	err := bi.bj.PlayerSplit()
	return bi.bjp.Output(bi.bj, err)
}

// Insurance インシュランス
func (bi *BlackJackInteractor) Insurance() string {
	err := bi.bj.PlayerInsurance()
	return bi.bjp.Output(bi.bj, err)
}

// DeclineInsurance インシュランス辞退
func (bi *BlackJackInteractor) DeclineInsurance() string {
	err := bi.bj.PlayerDeclineInsurance()
	return bi.bjp.Output(bi.bj, err)
}
