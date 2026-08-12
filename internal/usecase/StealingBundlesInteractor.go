//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// StealingBundlesInteractorIF スティーリングバンドルインタラクターインタフェース
type StealingBundlesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.StealingBundlesConfig) string
	// Take 場札を取る
	Take(cardIndex int) string
	// Steal 相手の束を奪う
	Steal(cardIndex, victimIdx int) string
	// Trail 場に置く
	Trail(cardIndex int) string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.StealingBundlesConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// StealingBundlesInteractor スティーリングバンドルインタラクタークラス
type StealingBundlesInteractor struct {
	GameBase[interfaces.StealingBundlesGame]
	sp presenter.StealingBundlesPresenter
}

// NewStealingBundlesInteractor コンストラクタ
func NewStealingBundlesInteractor(s interfaces.StealingBundlesGame, sp presenter.StealingBundlesPresenter) *StealingBundlesInteractor {
	mustNotNil("StealingBundlesInteractor", map[string]any{"s": s, "sp": sp})
	return &StealingBundlesInteractor{GameBase: GameBase[interfaces.StealingBundlesGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (si *StealingBundlesInteractor) Reset() string {
	si.Game.Reset()
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *StealingBundlesInteractor) ResetWithConfig(cfg domain.StealingBundlesConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Take 場札を取る
func (si *StealingBundlesInteractor) Take(cardIndex int) string {
	return si.act(func() error { return si.Game.PlayerTake(cardIndex) })
}

// Steal 相手の束を奪う
func (si *StealingBundlesInteractor) Steal(cardIndex, victimIdx int) string {
	return si.act(func() error { return si.Game.PlayerSteal(cardIndex, victimIdx) })
}

// Trail 場に置く
func (si *StealingBundlesInteractor) Trail(cardIndex int) string {
	return si.act(func() error { return si.Game.PlayerTrail(cardIndex) })
}

// act は 1 手打って CPU を進める共通処理。
func (si *StealingBundlesInteractor) act(play func() error) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := play(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// GiveUp 投了する
func (si *StealingBundlesInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.GiveUp()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *StealingBundlesInteractor) GetConfig() domain.StealingBundlesConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *StealingBundlesInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *StealingBundlesInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// advance は人間の出番が来るまでゲームを進める。
func (si *StealingBundlesInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if si.Game.GetGameEndFlag() || si.Game.IsHumanTurn() {
			return
		}
		si.Game.CpuPlay()
	}
}

// RestoreStealingBundlesInteractor deserialises JSON into a StealingBundlesInteractor.
func RestoreStealingBundlesInteractor(data []byte, sp presenter.StealingBundlesPresenter) (*StealingBundlesInteractor, error) {
	return restoreAndBuild[domain.StealingBundles](data, func(g *domain.StealingBundles) *StealingBundlesInteractor {
		return NewStealingBundlesInteractor(g, sp)
	})
}
