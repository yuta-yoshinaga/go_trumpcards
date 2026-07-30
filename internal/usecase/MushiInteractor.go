//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MushiInteractorIF 虫インタラクターインタフェース
type MushiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MushiConfig) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// Select 選択フェーズで取る場札を決める
	Select(fieldIdx int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MushiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MushiInteractor 虫インタラクタークラス
type MushiInteractor struct {
	GameBase[interfaces.MushiGame]
	mp presenter.MushiPresenter
}

// NewMushiInteractor コンストラクタ
func NewMushiInteractor(m interfaces.MushiGame, mp presenter.MushiPresenter) *MushiInteractor {
	mustNotNil("MushiInteractor", map[string]any{"m": m, "mp": mp})
	return &MushiInteractor{GameBase: GameBase[interfaces.MushiGame]{Game: m}, mp: mp}
}

// mushiHumanIdx 人間プレイヤーの座席。
const mushiHumanIdx = 0

// mushiCpuTurnCap は CPU ループの安全弁。1 ラウンドは 40 枚を高々 1 枚ずつ消費する
// ので、これを超えるのは停止条件が壊れているとき。無限ループより打ち切る。
const mushiCpuTurnCap = 128

// Reset ゲーム初期化
func (mi *MushiInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MushiInteractor) ResetWithConfig(cfg domain.MushiConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play 手札の札を出す
func (mi *MushiInteractor) Play(handIdx int) string {
	if mi.Game.GetGameEndFlag() {
		return mi.mp.Output(mi.Game, nil)
	}
	if err := mi.Game.PlayCard(mushiHumanIdx, handIdx); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// Select 選択フェーズで取る場札を決める
func (mi *MushiInteractor) Select(fieldIdx int) string {
	if mi.Game.GetGameEndFlag() {
		return mi.mp.Output(mi.Game, nil)
	}
	if err := mi.Game.SelectCapture(mushiHumanIdx, fieldIdx); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (mi *MushiInteractor) NextRound() string {
	if err := mi.Game.NextRound(); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MushiInteractor) GetConfig() domain.MushiConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *MushiInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MushiInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns 人間の手番に戻るかラウンドが終わるまで CPU を進める。
//
// 選択フェーズも CPU の番なら CPU が解決する。ここで戻ってしまうと、人間が
// 相手の選択を代わりに押すことになる。
func (mi *MushiInteractor) runCpuTurns() {
	for range mushiCpuTurnCap {
		if mi.Game.GetGameEndFlag() {
			return
		}
		phase := mi.Game.GetPhase()
		if phase == domain.MushiPhaseRoundEnd || phase == domain.MushiPhaseGameEnd {
			return
		}
		idx := mi.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == mushiHumanIdx {
			return
		}
		action := mi.Game.MushiCpuDecide(idx)
		var err error
		if phase == domain.MushiPhaseSelect || phase == domain.MushiPhaseWildSelect {
			err = mi.Game.SelectCapture(idx, action.FieldIdx)
		} else {
			err = mi.Game.PlayCard(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// RestoreMushiInteractor deserialises JSON into a MushiInteractor.
func RestoreMushiInteractor(data []byte, mp presenter.MushiPresenter) (*MushiInteractor, error) {
	return restoreAndBuild[domain.Mushi](data, func(g *domain.Mushi) *MushiInteractor {
		return &MushiInteractor{GameBase: GameBase[interfaces.MushiGame]{Game: g}, mp: mp}
	})
}
