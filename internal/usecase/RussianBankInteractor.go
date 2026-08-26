//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RussianBankInteractorIF ロシアンバンク (クラペット) のインタラクターインタフェース。
type RussianBankInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化。
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化。
	ResetWithConfig(cfg domain.RussianBankConfig) string
	// MoveToFoundation 移動元トップをファウンデーションへ移す。
	MoveToFoundation(zone int, fromOpp bool, col int) string
	// MoveToTableau 移動元トップを共有タブロー toCol へ移す。
	MoveToTableau(zone int, fromOpp bool, col, toCol int) string
	// Discard 手札を 1 枚捨てて手番終了。
	Discard() string
	// CallStop CPU の取りこぼしを咎める。
	CallStop() string
	// Undo 直近の人間の手を取り消す。
	Undo() string
	// GetConfig 現在の設定を取得。
	GetConfig() domain.RussianBankConfig
	// Hint ヒント取得。
	Hint() string
	// ActionLog 棋譜を出力する。
	ActionLog() string
}

// rbCpuGuard CPU 手番の自動進行ループ上限 (無限ループ防止)。
const rbCpuGuard = 100

// RussianBankInteractor ロシアンバンク (クラペット) のインタラクタークラス。
type RussianBankInteractor struct {
	GameBase[interfaces.RussianBankGame]
	sp presenter.RussianBankPresenter
}

// NewRussianBankInteractor コンストラクタ。
func NewRussianBankInteractor(g interfaces.RussianBankGame, sp presenter.RussianBankPresenter) *RussianBankInteractor {
	mustNotNil("RussianBankInteractor", map[string]any{"g": g, "sp": sp})
	return &RussianBankInteractor{GameBase: GameBase[interfaces.RussianBankGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化。
func (ti *RussianBankInteractor) Reset() string {
	ti.Game.Reset()
	return ti.sp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (ti *RussianBankInteractor) ResetWithConfig(cfg domain.RussianBankConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// MoveToFoundation 移動元トップをファウンデーションへ移す。
func (ti *RussianBankInteractor) MoveToFoundation(zone int, fromOpp bool, col int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.sp); blocked {
		return out
	}
	src := domain.RussianBankSource{Zone: domain.RussianBankZone(zone), FromOpponent: fromOpp, Col: col}
	if err := ti.Game.MoveToFoundation(src); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	return ti.sp.Output(ti.Game, nil)
}

// MoveToTableau 移動元トップを共有タブロー toCol へ移す。
func (ti *RussianBankInteractor) MoveToTableau(zone int, fromOpp bool, col, toCol int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.sp); blocked {
		return out
	}
	src := domain.RussianBankSource{Zone: domain.RussianBankZone(zone), FromOpponent: fromOpp, Col: col}
	if err := ti.Game.MoveToTableau(src, toCol); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	return ti.sp.Output(ti.Game, nil)
}

// Discard 手札を 1 枚捨てて手番終了し、CPU 手番を自動進行する。
func (ti *RussianBankInteractor) Discard() string {
	if out, blocked := guardNotPlayable(ti.Game, ti.sp); blocked {
		return out
	}
	if err := ti.Game.Discard(); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	ti.advanceCpu()
	return ti.sp.Output(ti.Game, nil)
}

// CallStop CPU の取りこぼしを咎める。
func (ti *RussianBankInteractor) CallStop() string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	if err := ti.Game.CallStop(); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	return ti.sp.Output(ti.Game, nil)
}

// Undo 直近の人間の手を取り消す。
func (ti *RussianBankInteractor) Undo() string {
	if err := ti.Game.Undo(); err != nil {
		return ti.sp.Output(ti.Game, err)
	}
	return ti.sp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得。
func (ti *RussianBankInteractor) GetConfig() domain.RussianBankConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得。
func (ti *RussianBankInteractor) Hint() string {
	return ti.sp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する。
func (ti *RussianBankInteractor) ActionLog() string {
	return ti.sp.ActionLogOutput(ti.Game)
}

// advanceCpu 手番が CPU の間、自動でターンを進める。
func (ti *RussianBankInteractor) advanceCpu() {
	for i := 0; i < rbCpuGuard; i++ {
		if ti.Game.GetGameEndFlag() || ti.Game.IsHumanTurn() {
			return
		}
		ti.Game.RunCpuTurn()
	}
}

// RestoreRussianBankInteractor deserialises JSON into a RussianBankInteractor.
func RestoreRussianBankInteractor(data []byte, sp presenter.RussianBankPresenter) (*RussianBankInteractor, error) {
	return restoreAndBuild[domain.RussianBank](data, func(g *domain.RussianBank) *RussianBankInteractor {
		return &RussianBankInteractor{GameBase: GameBase[interfaces.RussianBankGame]{Game: g}, sp: sp}
	})
}
