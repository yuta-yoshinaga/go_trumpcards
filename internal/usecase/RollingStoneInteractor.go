//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RollingStoneInteractorIF ローリングストーンインタラクターインタフェース
type RollingStoneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.RollingStoneConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// PickUp 場札を引き取る
	PickUp() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.RollingStoneConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RollingStoneInteractor ローリングストーンインタラクタークラス
type RollingStoneInteractor struct {
	GameBase[interfaces.RollingStoneGame]
	sp presenter.RollingStonePresenter
}

// NewRollingStoneInteractor コンストラクタ
func NewRollingStoneInteractor(s interfaces.RollingStoneGame, sp presenter.RollingStonePresenter) *RollingStoneInteractor {
	mustNotNil("RollingStoneInteractor", map[string]any{"s": s, "sp": sp})
	return &RollingStoneInteractor{GameBase: GameBase[interfaces.RollingStoneGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (ri *RollingStoneInteractor) Reset() string {
	ri.Game.Reset()
	ri.advance()
	return ri.sp.Output(ri.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ri *RollingStoneInteractor) ResetWithConfig(cfg domain.RollingStoneConfig) string {
	return resetWithValidatedConfig(ri.Game, ri.sp, cfg, ri.Game.SetConfig, ri.Reset)
}

// Play カードをプレイ
func (ri *RollingStoneInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ri.Game, ri.sp); blocked {
		return out
	}
	if err := ri.Game.PlayerPlay(cardIndex); err != nil {
		return ri.sp.Output(ri.Game, err)
	}
	ri.advance()
	return ri.sp.Output(ri.Game, nil)
}

// PickUp 場札を引き取る
func (ri *RollingStoneInteractor) PickUp() string {
	if out, blocked := guardNotPlayable(ri.Game, ri.sp); blocked {
		return out
	}
	if err := ri.Game.PlayerPickUp(); err != nil {
		return ri.sp.Output(ri.Game, err)
	}
	ri.advance()
	return ri.sp.Output(ri.Game, nil)
}

// GiveUp 投了する
func (ri *RollingStoneInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ri.Game, ri.sp); blocked {
		return out
	}
	ri.Game.GiveUp()
	return ri.sp.Output(ri.Game, nil)
}

// GetConfig 現在の設定を取得
func (ri *RollingStoneInteractor) GetConfig() domain.RollingStoneConfig { return ri.Game.GetConfig() }

// Hint ヒント取得
func (ri *RollingStoneInteractor) Hint() string { return ri.sp.HintOutput(ri.Game) }

// ActionLog 棋譜を出力する
func (ri *RollingStoneInteractor) ActionLog() string { return ri.sp.ActionLogOutput(ri.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **ラウンドの区切りが無い。** トリックは揃った時点で自動的に解決されるので、
// 途中で止まる段はプレイだけです。
func (ri *RollingStoneInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if ri.Game.GetGameEndFlag() || ri.Game.IsHumanTurn() {
			return
		}
		ri.Game.CpuPlay()
	}
}

// RestoreRollingStoneInteractor deserialises JSON into a RollingStoneInteractor.
func RestoreRollingStoneInteractor(data []byte, sp presenter.RollingStonePresenter) (*RollingStoneInteractor, error) {
	return restoreAndBuild[domain.RollingStone](data, func(g *domain.RollingStone) *RollingStoneInteractor {
		return NewRollingStoneInteractor(g, sp)
	})
}
