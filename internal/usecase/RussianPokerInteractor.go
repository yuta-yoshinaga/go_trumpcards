//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RussianPokerInteractorIF ロシアンポーカーインタラクターインタフェース
type RussianPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante int) string
	// Exchange カード交換（指定インデックス）
	Exchange(indices []int) string
	// Buy6th 6枚目のカードを購入する
	Buy6th() string
	// Select 6枚の手札から1枚を捨てて5枚にする
	Select(discardIndex int) string
	// Play プレイ（コール）
	Play() string
	// Fold フォールド
	Fold() string
	// ForceExchange ディーラーの最高カードを交換させる
	ForceExchange() string
	// Decline 強制クオリファイを辞退する
	Decline() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RussianPokerInteractor ロシアンポーカーインタラクタークラス
type RussianPokerInteractor struct {
	GameBase[interfaces.RussianPokerGame]
	rp presenter.RussianPokerPresenter
}

// NewRussianPokerInteractor コンストラクタ
func NewRussianPokerInteractor(g interfaces.RussianPokerGame, p presenter.RussianPokerPresenter) *RussianPokerInteractor {
	mustNotNil("RussianPokerInteractor", map[string]any{"g": g, "p": p})
	return &RussianPokerInteractor{
		GameBase: GameBase[interfaces.RussianPokerGame]{Game: g},
		rp:       p,
	}
}

// Reset ゲーム初期化
func (ri *RussianPokerInteractor) Reset() string {
	return runAndPresent(ri.Game, ri.rp, ri.Game.Reset)
}

// Bet アンテベット
func (ri *RussianPokerInteractor) Bet(ante int) string {
	return execAndPresent(ri.Game, ri.rp, func() error { return ri.Game.Bet(ante) })
}

// Exchange カード交換
func (ri *RussianPokerInteractor) Exchange(indices []int) string {
	return execAndPresent(ri.Game, ri.rp, func() error { return ri.Game.Exchange(indices) })
}

// Buy6th 6枚目のカードを購入
func (ri *RussianPokerInteractor) Buy6th() string {
	return execAndPresent(ri.Game, ri.rp, ri.Game.Buy6th)
}

// Select 6枚の手札から1枚を捨てる
func (ri *RussianPokerInteractor) Select(discardIndex int) string {
	return execAndPresent(ri.Game, ri.rp, func() error { return ri.Game.Select(discardIndex) })
}

// Play プレイ
func (ri *RussianPokerInteractor) Play() string {
	return execAndPresent(ri.Game, ri.rp, ri.Game.Play)
}

// Fold フォールド
func (ri *RussianPokerInteractor) Fold() string {
	return execAndPresent(ri.Game, ri.rp, ri.Game.Fold)
}

// ForceExchange ディーラーの最高カードを交換させる
func (ri *RussianPokerInteractor) ForceExchange() string {
	return execAndPresent(ri.Game, ri.rp, ri.Game.ForceExchange)
}

// Decline 強制クオリファイを辞退する
func (ri *RussianPokerInteractor) Decline() string {
	return execAndPresent(ri.Game, ri.rp, ri.Game.Decline)
}

// ActionLog 棋譜を出力する
func (ri *RussianPokerInteractor) ActionLog() string {
	return ri.rp.ActionLogOutput(ri.Game)
}

// RestoreRussianPokerInteractor deserialises JSON into a RussianPokerInteractor.
func RestoreRussianPokerInteractor(data []byte, p presenter.RussianPokerPresenter) (*RussianPokerInteractor, error) {
	return restoreAndBuild[domain.RussianPoker](data, func(g *domain.RussianPoker) *RussianPokerInteractor {
		return &RussianPokerInteractor{GameBase: GameBase[interfaces.RussianPokerGame]{Game: g}, rp: p}
	})
}
