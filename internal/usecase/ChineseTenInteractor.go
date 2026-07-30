//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ChineseTenInteractorIF 撿紅點インタラクターインタフェース
type ChineseTenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ChineseTenConfig) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// Select 選択フェーズで取る場札を決める
	Select(layoutIdx int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ChineseTenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ChineseTenInteractor 撿紅點インタラクタークラス
type ChineseTenInteractor struct {
	GameBase[interfaces.ChineseTenGame]
	cp presenter.ChineseTenPresenter
}

// NewChineseTenInteractor コンストラクタ
func NewChineseTenInteractor(c interfaces.ChineseTenGame, cp presenter.ChineseTenPresenter) *ChineseTenInteractor {
	mustNotNil("ChineseTenInteractor", map[string]any{"c": c, "cp": cp})
	return &ChineseTenInteractor{GameBase: GameBase[interfaces.ChineseTenGame]{Game: c}, cp: cp}
}

// chineseTenHumanIdx 人間プレイヤーの座席。
const chineseTenHumanIdx = 0

// chineseTenCpuTurnCap は CPU ループの安全弁。1 ゲームは 24 手番を超えないので、
// これを超えるのは停止条件が壊れているとき。
const chineseTenCpuTurnCap = 128

// Reset ゲーム初期化
func (ci *ChineseTenInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ChineseTenInteractor) ResetWithConfig(cfg domain.ChineseTenConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play 手札の札を出す
func (ci *ChineseTenInteractor) Play(handIdx int) string {
	if ci.Game.GetGameEndFlag() {
		return ci.cp.Output(ci.Game, nil)
	}
	if err := ci.Game.PlayCard(chineseTenHumanIdx, handIdx); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// Select 選択フェーズで取る場札を決める
func (ci *ChineseTenInteractor) Select(layoutIdx int) string {
	if ci.Game.GetGameEndFlag() {
		return ci.cp.Output(ci.Game, nil)
	}
	if err := ci.Game.SelectCapture(chineseTenHumanIdx, layoutIdx); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ChineseTenInteractor) GetConfig() domain.ChineseTenConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *ChineseTenInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *ChineseTenInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// 選択フェーズも CPU の番なら CPU が解決する。ここで戻ると、人間が相手の選択を
// 代わりに押すことになる。
func (ci *ChineseTenInteractor) runCpuTurns() {
	for range chineseTenCpuTurnCap {
		if ci.Game.GetGameEndFlag() {
			return
		}
		idx := ci.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == chineseTenHumanIdx {
			return
		}
		action := ci.Game.ChineseTenCpuDecide(idx)
		var err error
		if ci.Game.GetPhase() == domain.ChineseTenPhaseSelect {
			err = ci.Game.SelectCapture(idx, action.LayoutIdx)
		} else {
			err = ci.Game.PlayCard(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// RestoreChineseTenInteractor deserialises JSON into a ChineseTenInteractor.
func RestoreChineseTenInteractor(data []byte, cp presenter.ChineseTenPresenter) (*ChineseTenInteractor, error) {
	return restoreAndBuild[domain.ChineseTen](data, func(g *domain.ChineseTen) *ChineseTenInteractor {
		return &ChineseTenInteractor{GameBase: GameBase[interfaces.ChineseTenGame]{Game: g}, cp: cp}
	})
}
