//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SkitgubbeInteractorIF シートグッベインタラクターインタフェース
type SkitgubbeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SkitgubbeConfig) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// PickUp 第2フェーズで場の札を引き取る
	PickUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SkitgubbeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SkitgubbeInteractor シートグッベインタラクタークラス
type SkitgubbeInteractor struct {
	GameBase[interfaces.SkitgubbeGame]
	cp presenter.SkitgubbePresenter
}

// NewSkitgubbeInteractor コンストラクタ
func NewSkitgubbeInteractor(c interfaces.SkitgubbeGame, cp presenter.SkitgubbePresenter) *SkitgubbeInteractor {
	mustNotNil("SkitgubbeInteractor", map[string]any{"c": c, "cp": cp})
	return &SkitgubbeInteractor{GameBase: GameBase[interfaces.SkitgubbeGame]{Game: c}, cp: cp}
}

// skitgubbeHumanIdx 人間プレイヤーの座席。
const skitgubbeHumanIdx = 0

// skitgubbeCpuTurnCap は CPU ループの安全弁。第2フェーズは引き取りで手札が戻る
// ので長引くが、それでもこの回数には届かない。超えるのは停止条件が壊れたとき。
const skitgubbeCpuTurnCap = 512

// Reset ゲーム初期化
func (si *SkitgubbeInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SkitgubbeInteractor) ResetWithConfig(cfg domain.SkitgubbeConfig) string {
	return resetWithValidatedConfig(si.Game, si.cp, cfg, si.Game.SetConfig, si.Reset)
}

// Play 手札の札を出す
func (si *SkitgubbeInteractor) Play(handIdx int) string {
	if si.Game.GetGameEndFlag() {
		return si.cp.Output(si.Game, nil)
	}
	if err := si.Game.PlayCard(skitgubbeHumanIdx, handIdx); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// PickUp 第2フェーズで場の札を引き取る
func (si *SkitgubbeInteractor) PickUp() string {
	if si.Game.GetGameEndFlag() {
		return si.cp.Output(si.Game, nil)
	}
	if err := si.Game.PickUp(skitgubbeHumanIdx); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SkitgubbeInteractor) GetConfig() domain.SkitgubbeConfig { return si.Game.GetConfig() }

// Hint ヒント取得
func (si *SkitgubbeInteractor) Hint() string { return si.cp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *SkitgubbeInteractor) ActionLog() string { return si.cp.ActionLogOutput(si.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
func (si *SkitgubbeInteractor) runCpuTurns() {
	for range skitgubbeCpuTurnCap {
		if si.Game.GetGameEndFlag() {
			return
		}
		idx := si.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == skitgubbeHumanIdx {
			return
		}
		action := si.Game.SkitgubbeCpuDecide(idx)
		var err error
		if action.PickUp {
			err = si.Game.PickUp(idx)
		} else {
			err = si.Game.PlayCard(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// RestoreSkitgubbeInteractor deserialises JSON into a SkitgubbeInteractor.
func RestoreSkitgubbeInteractor(data []byte, cp presenter.SkitgubbePresenter) (*SkitgubbeInteractor, error) {
	return restoreAndBuild[domain.Skitgubbe](data, func(g *domain.Skitgubbe) *SkitgubbeInteractor {
		return &SkitgubbeInteractor{GameBase: GameBase[interfaces.SkitgubbeGame]{Game: g}, cp: cp}
	})
}
