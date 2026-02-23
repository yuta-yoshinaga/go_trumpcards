package usecases

import (
	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/usecases/presenters"
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
	p  *entities.Poker
	pp presenters.PokerPresenter
}

// NewPokerInteractor コンストラクタ
func NewPokerInteractor(pp presenters.PokerPresenter) *PokerInteractor {
	return &PokerInteractor{
		p:  entities.NewPoker(entities.NewTrumpCards(0), entities.NewPokerPlayer(), entities.NewPokerPlayer()),
		pp: pp,
	}
}

// Reset ゲーム初期化
func (pi *PokerInteractor) Reset() string {
	pi.p.Reset()
	return pi.pp.Output(pi.p)
}

// Exchange カード交換
func (pi *PokerInteractor) Exchange(indices []int) string {
	pi.p.PlayerExchange(indices)
	return pi.pp.Output(pi.p)
}

// Stand カード交換なしでショーダウン
func (pi *PokerInteractor) Stand() string {
	pi.p.PlayerStand()
	return pi.pp.Output(pi.p)
}

// Bet ベット
func (pi *PokerInteractor) Bet(amount int) string {
	pi.p.PlayerBet(amount)
	return pi.pp.Output(pi.p)
}

// Call コール
func (pi *PokerInteractor) Call() string {
	pi.p.PlayerCall()
	return pi.pp.Output(pi.p)
}

// Raise レイズ
func (pi *PokerInteractor) Raise(amount int) string {
	pi.p.PlayerRaise(amount)
	return pi.pp.Output(pi.p)
}

// Fold フォールド
func (pi *PokerInteractor) Fold() string {
	pi.p.PlayerFold()
	return pi.pp.Output(pi.p)
}

// Check チェック
func (pi *PokerInteractor) Check() string {
	pi.p.PlayerCheck()
	return pi.pp.Output(pi.p)
}
