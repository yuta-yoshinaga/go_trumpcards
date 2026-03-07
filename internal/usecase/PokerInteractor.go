package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PokerInteractorIF ポーカーインタラクターインタフェース
type PokerInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.PokerConfig) string
	Action(action int, amount int) string
	Exchange(indices []int) string
	Stand() string
	Odds(indices []int) string
}

// PokerInteractor ポーカーインタラクタークラス
type PokerInteractor struct {
	p  interfaces.PokerGame
	pp presenter.PokerPresenter
}

// NewPokerInteractor コンストラクタ
func NewPokerInteractor(p interfaces.PokerGame, pp presenter.PokerPresenter) *PokerInteractor {
	mustNotNil("PokerInteractor", map[string]any{"p": p, "pp": pp})
	return &PokerInteractor{
		p:  p,
		pp: pp,
	}
}

// Reset ゲーム初期化
func (pi *PokerInteractor) Reset() string {
	err := pi.p.Reset()
	return pi.pp.Output(pi.p, err)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PokerInteractor) ResetWithConfig(cfg domain.PokerConfig) string {
	pi.p.SetConfig(cfg)
	err := pi.p.Reset()
	return pi.pp.Output(pi.p, err)
}

// Action プレイヤーアクション実行
func (pi *PokerInteractor) Action(action int, amount int) string {
	err := pi.p.PlayerAction(action, amount)
	return pi.pp.Output(pi.p, err)
}

// Exchange カード交換
func (pi *PokerInteractor) Exchange(indices []int) string {
	err := pi.p.PlayerExchange(indices)
	return pi.pp.Output(pi.p, err)
}

// Stand カード交換なし
func (pi *PokerInteractor) Stand() string {
	err := pi.p.PlayerStand()
	return pi.pp.Output(pi.p, err)
}

// Odds ドローオッズ計算
func (pi *PokerInteractor) Odds(indices []int) string {
	odds, err := pi.p.CalcDrawOdds(indices)
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	return pi.pp.OutputWithOdds(pi.p, nil, odds)
}
