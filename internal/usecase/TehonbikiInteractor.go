//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TehonbikiInteractorIF 手本引きインタラクターインタフェース
type TehonbikiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TehonbikiConfig) string
	// PlaceBet 場札に賭けてゲートをめくる
	PlaceBet(numbers []int, kind domain.TehonbikiBetType, bet int) string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TehonbikiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TehonbikiInteractor 手本引きインタラクタークラス
type TehonbikiInteractor struct {
	GameBase[interfaces.TehonbikiGame]
	cp presenter.TehonbikiPresenter
}

// NewTehonbikiInteractor コンストラクタ
func NewTehonbikiInteractor(c interfaces.TehonbikiGame, cp presenter.TehonbikiPresenter) *TehonbikiInteractor {
	mustNotNil("TehonbikiInteractor", map[string]any{"c": c, "cp": cp})
	return &TehonbikiInteractor{GameBase: GameBase[interfaces.TehonbikiGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *TehonbikiInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *TehonbikiInteractor) ResetWithConfig(cfg domain.TehonbikiConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet 場札に賭けてゲートをめくる
//
// **どの場札が得かはここで判定しない。** 場札に何枚同じスートが出ているかで
// 期待値が変わるが、それはドメインの規則であって、ここで賭けを拒んだり
// 選び直させたりする話ではない ── 損な賭けも、プレイヤーが選べる手である。
func (ci *TehonbikiInteractor) PlaceBet(numbers []int, kind domain.TehonbikiBetType, bet int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(numbers, kind, bet) })
}

// NextRound 次のラウンドを始める
func (ci *TehonbikiInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
func (ci *TehonbikiInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *TehonbikiInteractor) GetConfig() domain.TehonbikiConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *TehonbikiInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *TehonbikiInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreTehonbikiInteractor deserialises JSON into an interactor.
func RestoreTehonbikiInteractor(data []byte, cp presenter.TehonbikiPresenter) (*TehonbikiInteractor, error) {
	return restoreAndBuild[domain.Tehonbiki](data,
		func(g *domain.Tehonbiki) *TehonbikiInteractor { return NewTehonbikiInteractor(g, cp) })
}
