//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OasisPokerInteractorIF オアシスポーカーインタラクターインタフェース
type OasisPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, jackpot int) string
	// Exchange カード交換（指定インデックス）
	Exchange(indices []int) string
	// Stand 交換せずアクションフェーズへ進む
	Stand() string
	// Play プレイ（コール）
	Play() string
	// Fold フォールド
	Fold() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// OasisPokerInteractor オアシスポーカーインタラクタークラス
type OasisPokerInteractor struct {
	GameBase[interfaces.OasisPokerGame]
	op presenter.OasisPokerPresenter
}

// NewOasisPokerInteractor コンストラクタ
func NewOasisPokerInteractor(g interfaces.OasisPokerGame, p presenter.OasisPokerPresenter) *OasisPokerInteractor {
	mustNotNil("OasisPokerInteractor", map[string]any{"g": g, "p": p})
	return &OasisPokerInteractor{
		GameBase: GameBase[interfaces.OasisPokerGame]{Game: g},
		op:       p,
	}
}

// Reset ゲーム初期化
func (oi *OasisPokerInteractor) Reset() string {
	return runAndPresent(oi.Game, oi.op, oi.Game.Reset)
}

// Bet アンテベット
func (oi *OasisPokerInteractor) Bet(ante, jackpot int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.Bet(ante, jackpot) })
}

// Exchange カード交換
func (oi *OasisPokerInteractor) Exchange(indices []int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.Exchange(indices) })
}

// Stand 交換せずアクションフェーズへ進む
func (oi *OasisPokerInteractor) Stand() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Stand)
}

// Play プレイ
func (oi *OasisPokerInteractor) Play() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Play)
}

// Fold フォールド
func (oi *OasisPokerInteractor) Fold() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Fold)
}

// ActionLog 棋譜を出力する
func (oi *OasisPokerInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// Hint ヒントを出力する
func (oi *OasisPokerInteractor) Hint() string {
	return oi.op.HintOutput(oi.Game)
}

// RestoreOasisPokerInteractor deserialises JSON into an OasisPokerInteractor.
func RestoreOasisPokerInteractor(data []byte, p presenter.OasisPokerPresenter) (*OasisPokerInteractor, error) {
	return restoreAndBuild[domain.OasisPoker](data, func(g *domain.OasisPoker) *OasisPokerInteractor {
		return &OasisPokerInteractor{GameBase: GameBase[interfaces.OasisPokerGame]{Game: g}, op: p}
	})
}
