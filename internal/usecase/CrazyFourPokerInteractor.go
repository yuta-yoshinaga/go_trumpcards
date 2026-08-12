//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CrazyFourPokerInteractorIF クレイジー 4 ポーカーインタラクターインタフェース
type CrazyFourPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CrazyFourPokerConfig) string
	// PlaceBet アンティと任意の Queens Up を置く
	PlaceBet(ante, queensUp int) string
	// Play プレイベットを置く
	Play(multiplier int) string
	// Fold 降りる
	Fold() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CrazyFourPokerConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CrazyFourPokerInteractor クレイジー 4 ポーカーインタラクタークラス
type CrazyFourPokerInteractor struct {
	GameBase[interfaces.CrazyFourPokerGame]
	cp presenter.CrazyFourPokerPresenter
}

// NewCrazyFourPokerInteractor コンストラクタ
func NewCrazyFourPokerInteractor(c interfaces.CrazyFourPokerGame, cp presenter.CrazyFourPokerPresenter) *CrazyFourPokerInteractor {
	mustNotNil("CrazyFourPokerInteractor", map[string]any{"c": c, "cp": cp})
	return &CrazyFourPokerInteractor{GameBase: GameBase[interfaces.CrazyFourPokerGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *CrazyFourPokerInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CrazyFourPokerInteractor) ResetWithConfig(cfg domain.CrazyFourPokerConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet アンティと任意の Queens Up を置く
func (ci *CrazyFourPokerInteractor) PlaceBet(ante, queensUp int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(ante, queensUp) })
}

// Play プレイベットを置く
func (ci *CrazyFourPokerInteractor) Play(multiplier int) string {
	return ci.runGuarded(func() error { return ci.Game.Play(multiplier) })
}

// Fold 降りる
func (ci *CrazyFourPokerInteractor) Fold() string {
	return ci.runGuarded(ci.Game.Fold)
}

// NextRound 次のラウンドを始める
func (ci *CrazyFourPokerInteractor) NextRound() string {
	return ci.runGuarded(ci.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **上限倍率の判定はドメインに任せます。** ここで手役を見て 3 倍を通すかどうかを
// 決め直すと、規則が 2 か所に増えて必ずずれます。
func (ci *CrazyFourPokerInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CrazyFourPokerInteractor) GetConfig() domain.CrazyFourPokerConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *CrazyFourPokerInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *CrazyFourPokerInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreCrazyFourPokerInteractor deserialises JSON into a CrazyFourPokerInteractor.
func RestoreCrazyFourPokerInteractor(data []byte, cp presenter.CrazyFourPokerPresenter) (*CrazyFourPokerInteractor, error) {
	return restoreAndBuild[domain.CrazyFourPoker](data, func(g *domain.CrazyFourPoker) *CrazyFourPokerInteractor {
		return NewCrazyFourPokerInteractor(g, cp)
	})
}
