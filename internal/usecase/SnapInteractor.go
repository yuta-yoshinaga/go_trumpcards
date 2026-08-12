//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SnapInteractorIF スナップインタラクターインタフェース
type SnapInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SnapConfig) string
	// Step 1枚めくる
	Step() string
	// Snap スナップを宣言する
	Snap() string
	// Tick 保留中の CPU アクションを進める
	Tick() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SnapConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SnapInteractor スナップインタラクタークラス
type SnapInteractor struct {
	GameBase[interfaces.SnapGame]
	sp presenter.SnapPresenter
}

// NewSnapInteractor コンストラクタ
func NewSnapInteractor(s interfaces.SnapGame, sp presenter.SnapPresenter) *SnapInteractor {
	mustNotNil("SnapInteractor", map[string]any{"s": s, "sp": sp})
	return &SnapInteractor{GameBase: GameBase[interfaces.SnapGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SnapInteractor) Reset() string {
	si.Game.Reset()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SnapInteractor) ResetWithConfig(cfg domain.SnapConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Step 1枚めくる
func (si *SnapInteractor) Step() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerStep(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	return si.sp.Output(si.Game, nil)
}

// Snap スナップを宣言する
//
// **CPU を先に進めない。** ここで Tick を回すと、人間の宣言より前に CPU の
// 予約が発火してしまい、反射ゲームとして成立しません。
func (si *SnapInteractor) Snap() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerSnap(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	return si.sp.Output(si.Game, nil)
}

// Tick 保留中の CPU アクションを進める
func (si *SnapInteractor) Tick() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.Tick()
	return si.sp.Output(si.Game, nil)
}

// GiveUp 投了する
func (si *SnapInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.GiveUp()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SnapInteractor) GetConfig() domain.SnapConfig { return si.Game.GetConfig() }

// Hint ヒント取得
func (si *SnapInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *SnapInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// RestoreSnapInteractor deserialises JSON into a SnapInteractor.
func RestoreSnapInteractor(data []byte, sp presenter.SnapPresenter) (*SnapInteractor, error) {
	return restoreAndBuild[domain.Snap](data, func(g *domain.Snap) *SnapInteractor {
		return NewSnapInteractor(g, sp)
	})
}
