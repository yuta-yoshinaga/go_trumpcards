package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PokerInteractorIF ポーカーインタラクターインタフェース
type PokerInteractorIF interface {
	Reset() string
	Exchange(indices []int) string
	Stand() string
	Bet(amount int) string
	Call() string
	Raise(amount int) string
	Fold() string
	Check() string
}

// PokerInteractor ポーカーインタラクタークラス
type PokerInteractor struct {
	p  interfaces.PokerGame
	pp presenter.PokerPresenter
}

// NewPokerInteractor コンストラクタ
func NewPokerInteractor(p interfaces.PokerGame, pp presenter.PokerPresenter) *PokerInteractor {
	if p == nil {
		panic("PokerInteractor: p must not be nil")
	}
	if pp == nil {
		panic("PokerInteractor: pp must not be nil")
	}
	return &PokerInteractor{
		p:  p,
		pp: pp,
	}
}

// Reset ゲーム初期化
func (pi *PokerInteractor) Reset() string {
	pi.p.Reset()
	return pi.pp.Output(pi.p, nil)
}

// Exchange カード交換
func (pi *PokerInteractor) Exchange(indices []int) string {
	err := pi.p.PlayerExchange(indices)
	return pi.pp.Output(pi.p, err)
}

// Stand カード交換なしでショーダウン
func (pi *PokerInteractor) Stand() string {
	err := pi.p.PlayerStand()
	return pi.pp.Output(pi.p, err)
}

// Bet ベット
func (pi *PokerInteractor) Bet(amount int) string {
	err := pi.p.PlayerBet(amount)
	return pi.pp.Output(pi.p, err)
}

// Call コール
func (pi *PokerInteractor) Call() string {
	err := pi.p.PlayerCall()
	return pi.pp.Output(pi.p, err)
}

// Raise レイズ
func (pi *PokerInteractor) Raise(amount int) string {
	err := pi.p.PlayerRaise(amount)
	return pi.pp.Output(pi.p, err)
}

// Fold フォールド
func (pi *PokerInteractor) Fold() string {
	err := pi.p.PlayerFold()
	return pi.pp.Output(pi.p, err)
}

// Check チェック
func (pi *PokerInteractor) Check() string {
	err := pi.p.PlayerCheck()
	return pi.pp.Output(pi.p, err)
}
