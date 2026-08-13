//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BaseballPokerInteractorIF ベースボールポーカーインタラクターインタフェース
type BaseballPokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BaseballPokerConfig) string
	// Action 人間の手を処理する
	Action(action, amount int) string
	// AnswerBuyIn 買い増しの返事を処理する
	AnswerBuyIn(answer int) string
	// NextHand 次のハンドを始める
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BaseballPokerConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BaseballPokerInteractor ベースボールポーカーインタラクタークラス
type BaseballPokerInteractor struct {
	GameBase[interfaces.BaseballPokerGame]
	cp presenter.BaseballPokerPresenter
}

// NewBaseballPokerInteractor コンストラクタ
func NewBaseballPokerInteractor(c interfaces.BaseballPokerGame, cp presenter.BaseballPokerPresenter) *BaseballPokerInteractor {
	mustNotNil("BaseballPokerInteractor", map[string]any{"c": c, "cp": cp})
	return &BaseballPokerInteractor{GameBase: GameBase[interfaces.BaseballPokerGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *BaseballPokerInteractor) Reset() string {
	ci.Game.Reset()
	// **配った直後から CPU が動くことがある。** 表の 3 で買い増しを迫られるのが
	// CPU なら、その返事まで進めないと人間の手番に戻らない。
	ci.Game.CpuPlay()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *BaseballPokerInteractor) ResetWithConfig(cfg domain.BaseballPokerConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Action 人間の手を処理する
func (ci *BaseballPokerInteractor) Action(action, amount int) string {
	return ci.runGuarded(func() error { return ci.Game.PlayerAction(action, amount) })
}

// AnswerBuyIn は買い増しの返事を処理する。
//
// **払うか降りるかはプレイヤーが決める。** 手が強いからと勝手に払わせると、
// このゲームでいちばん高くつく判断が消える。
func (ci *BaseballPokerInteractor) AnswerBuyIn(answer int) string {
	return ci.runGuarded(func() error { return ci.Game.AnswerBuyIn(answer) })
}

// NextHand 次のハンドを始める
func (ci *BaseballPokerInteractor) NextHand() string { return ci.runGuarded(ci.Game.NextHand) }

// runGuarded は終局後の操作を弾いてから action を実行し、CPU を進めて出力する。
func (ci *BaseballPokerInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.Game.CpuPlay()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *BaseballPokerInteractor) GetConfig() domain.BaseballPokerConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *BaseballPokerInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *BaseballPokerInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreBaseballPokerInteractor deserialises JSON into an interactor.
func RestoreBaseballPokerInteractor(data []byte, cp presenter.BaseballPokerPresenter) (*BaseballPokerInteractor, error) {
	return restoreAndBuild[domain.BaseballPoker](data,
		func(g *domain.BaseballPoker) *BaseballPokerInteractor {
			return NewBaseballPokerInteractor(g, cp)
		})
}
