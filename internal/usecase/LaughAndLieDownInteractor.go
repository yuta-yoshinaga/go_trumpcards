//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LaughAndLieDownInteractorIF ラフ・アンド・ライダウンインタラクターインタフェース
type LaughAndLieDownInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.LaughAndLieDownConfig) string
	// Play 手札の札を出し、場から takeCount 枚を取る
	Play(handIdx, takeCount int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.LaughAndLieDownConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LaughAndLieDownInteractor ラフ・アンド・ライダウンインタラクタークラス
type LaughAndLieDownInteractor struct {
	GameBase[interfaces.LaughAndLieDownGame]
	cp presenter.LaughAndLieDownPresenter
}

// NewLaughAndLieDownInteractor コンストラクタ
func NewLaughAndLieDownInteractor(c interfaces.LaughAndLieDownGame, cp presenter.LaughAndLieDownPresenter) *LaughAndLieDownInteractor {
	mustNotNil("LaughAndLieDownInteractor", map[string]any{"c": c, "cp": cp})
	return &LaughAndLieDownInteractor{GameBase: GameBase[interfaces.LaughAndLieDownGame]{Game: c}, cp: cp}
}

// laughAndLieDownHumanIdx 人間プレイヤーの座席。
const laughAndLieDownHumanIdx = 0

// laughAndLieDownCpuTurnCap は CPU ループの安全弁。1 ゲームは 40 手番を超えないので、
// これを超えるのは停止条件が壊れているとき。
const laughAndLieDownCpuTurnCap = 128

// Reset ゲーム初期化
func (li *LaughAndLieDownInteractor) Reset() string {
	li.Game.Reset()
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (li *LaughAndLieDownInteractor) ResetWithConfig(cfg domain.LaughAndLieDownConfig) string {
	return resetWithValidatedConfig(li.Game, li.cp, cfg, li.Game.SetConfig, li.Reset)
}

// Play 手札の札を出し、場から takeCount 枚を取る
func (li *LaughAndLieDownInteractor) Play(handIdx, takeCount int) string {
	if li.Game.GetGameEndFlag() {
		return li.cp.Output(li.Game, nil)
	}
	if err := li.Game.PlayCard(laughAndLieDownHumanIdx, handIdx, takeCount); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// GetConfig 現在の設定を取得
func (li *LaughAndLieDownInteractor) GetConfig() domain.LaughAndLieDownConfig {
	return li.Game.GetConfig()
}

// Hint ヒント取得
func (li *LaughAndLieDownInteractor) Hint() string { return li.cp.HintOutput(li.Game) }

// ActionLog 棋譜を出力する
func (li *LaughAndLieDownInteractor) ActionLog() string { return li.cp.ActionLogOutput(li.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// 「取れなければ降りる」はドメインが手番送りの中で解決するので、ここは出す手だけを
// 送ればよい。
func (li *LaughAndLieDownInteractor) runCpuTurns() {
	for range laughAndLieDownCpuTurnCap {
		if li.Game.GetGameEndFlag() {
			return
		}
		idx := li.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == laughAndLieDownHumanIdx {
			return
		}
		action := li.Game.LaughAndLieDownCpuDecide(idx)
		if action.HandIdx < 0 {
			return
		}
		if err := li.Game.PlayCard(idx, action.HandIdx, action.TakeCount); err != nil {
			return
		}
	}
}

// RestoreLaughAndLieDownInteractor deserialises JSON into a LaughAndLieDownInteractor.
func RestoreLaughAndLieDownInteractor(data []byte, cp presenter.LaughAndLieDownPresenter) (*LaughAndLieDownInteractor, error) {
	return restoreAndBuild[domain.LaughAndLieDown](data, func(g *domain.LaughAndLieDown) *LaughAndLieDownInteractor {
		return &LaughAndLieDownInteractor{GameBase: GameBase[interfaces.LaughAndLieDownGame]{Game: g}, cp: cp}
	})
}
