//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// IronCrossInteractorIF アイアンクロスインタラクターインタフェース
type IronCrossInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.IronCrossConfig) string
	// Action 人間の手を処理する
	Action(action, amount int) string
	// ChooseLine 使う列を決める
	ChooseLine(line int) string
	// NextHand 次のハンドを始める
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.IronCrossConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// IronCrossInteractor アイアンクロスインタラクタークラス
type IronCrossInteractor struct {
	GameBase[interfaces.IronCrossGame]
	cp presenter.IronCrossPresenter
}

// NewIronCrossInteractor コンストラクタ
func NewIronCrossInteractor(c interfaces.IronCrossGame, cp presenter.IronCrossPresenter) *IronCrossInteractor {
	mustNotNil("IronCrossInteractor", map[string]any{"c": c, "cp": cp})
	return &IronCrossInteractor{GameBase: GameBase[interfaces.IronCrossGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *IronCrossInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *IronCrossInteractor) ResetWithConfig(cfg domain.IronCrossConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Action 人間の手を処理する
func (ci *IronCrossInteractor) Action(action, amount int) string {
	return ci.runGuarded(func() error { return ci.Game.PlayerAction(action, amount) })
}

// ChooseLine は使う列を決める。
//
// **どちらが強いかはここで判定しない。** 弱いほうを選ぶのもプレイヤーの手で、
// 良かれと思って強いほうへ直すと、このゲームの唯一の判断が消える。
func (ci *IronCrossInteractor) ChooseLine(line int) string {
	return ci.runGuarded(func() error { return ci.Game.ChooseLine(domain.IronCrossLine(line)) })
}

// NextHand 次のハンドを始める
func (ci *IronCrossInteractor) NextHand() string { return ci.runGuarded(ci.Game.NextHand) }

// runGuarded は終局後の操作を弾いてから action を実行し、CPU を進めて出力する。
func (ci *IronCrossInteractor) runGuarded(action func() error) string {
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
func (ci *IronCrossInteractor) GetConfig() domain.IronCrossConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *IronCrossInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *IronCrossInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreIronCrossInteractor deserialises JSON into an interactor.
func RestoreIronCrossInteractor(data []byte, cp presenter.IronCrossPresenter) (*IronCrossInteractor, error) {
	return restoreAndBuild[domain.IronCross](data,
		func(g *domain.IronCross) *IronCrossInteractor { return NewIronCrossInteractor(g, cp) })
}
