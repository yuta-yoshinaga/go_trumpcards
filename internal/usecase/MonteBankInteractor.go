//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MonteBankInteractorIF モンテバンクインタラクターインタフェース
type MonteBankInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MonteBankConfig) string
	// PlaceBet 場札に賭けてゲートをめくる
	PlaceBet(idx, bet int) string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MonteBankConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MonteBankInteractor モンテバンクインタラクタークラス
type MonteBankInteractor struct {
	GameBase[interfaces.MonteBankGame]
	cp presenter.MonteBankPresenter
}

// NewMonteBankInteractor コンストラクタ
func NewMonteBankInteractor(c interfaces.MonteBankGame, cp presenter.MonteBankPresenter) *MonteBankInteractor {
	mustNotNil("MonteBankInteractor", map[string]any{"c": c, "cp": cp})
	return &MonteBankInteractor{GameBase: GameBase[interfaces.MonteBankGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *MonteBankInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *MonteBankInteractor) ResetWithConfig(cfg domain.MonteBankConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet 場札に賭けてゲートをめくる
//
// **どの場札が得かはここで判定しない。** 場札に何枚同じスートが出ているかで
// 期待値が変わるが、それはドメインの規則であって、ここで賭けを拒んだり
// 選び直させたりする話ではない ── 損な賭けも、プレイヤーが選べる手である。
func (ci *MonteBankInteractor) PlaceBet(idx, bet int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(idx, bet) })
}

// NextRound 次のラウンドを始める
func (ci *MonteBankInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
func (ci *MonteBankInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *MonteBankInteractor) GetConfig() domain.MonteBankConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *MonteBankInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *MonteBankInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreMonteBankInteractor deserialises JSON into an interactor.
func RestoreMonteBankInteractor(data []byte, cp presenter.MonteBankPresenter) (*MonteBankInteractor, error) {
	return restoreAndBuild[domain.MonteBank](data,
		func(g *domain.MonteBank) *MonteBankInteractor { return NewMonteBankInteractor(g, cp) })
}
