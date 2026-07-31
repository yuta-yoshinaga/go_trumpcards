//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PopeJoanInteractorIF ポープ・ジョーンインタラクターインタフェース
type PopeJoanInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PopeJoanConfig) string
	// Play 手札1枚を出す
	Play(handIdx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PopeJoanConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PopeJoanInteractor ポープ・ジョーンインタラクタークラス
type PopeJoanInteractor struct {
	GameBase[interfaces.PopeJoanGame]
	cp presenter.PopeJoanPresenter
}

// NewPopeJoanInteractor コンストラクタ
func NewPopeJoanInteractor(c interfaces.PopeJoanGame, cp presenter.PopeJoanPresenter) *PopeJoanInteractor {
	mustNotNil("PopeJoanInteractor", map[string]any{"c": c, "cp": cp})
	return &PopeJoanInteractor{GameBase: GameBase[interfaces.PopeJoanGame]{Game: c}, cp: cp}
}

// popeJoanHumanIdx 人間プレイヤーの座席。
const popeJoanHumanIdx = 0

// popeJoanCpuTurnCap は CPU ループの安全弁。
const popeJoanCpuTurnCap = 2000

// Reset ゲーム初期化
func (pi *PopeJoanInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PopeJoanInteractor) ResetWithConfig(cfg domain.PopeJoanConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.cp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Play 手札1枚を出す
func (pi *PopeJoanInteractor) Play(handIdx int) string {
	if pi.Game.GetGameEndFlag() {
		return pi.cp.Output(pi.Game, nil)
	}
	if err := pi.Game.Play(popeJoanHumanIdx, handIdx); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// NextDeal 次のディールへ進む
func (pi *PopeJoanInteractor) NextDeal() string {
	if err := pi.Game.NextDeal(); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PopeJoanInteractor) GetConfig() domain.PopeJoanConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *PopeJoanInteractor) Hint() string { return pi.cp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *PopeJoanInteractor) ActionLog() string { return pi.cp.ActionLogOutput(pi.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ディール終了では止める。勝手に配り直すと、8 区画の精算を読む間もなく画面が
// 変わってしまう。
func (pi *PopeJoanInteractor) runCpuTurns() {
	for range popeJoanCpuTurnCap {
		if pi.Game.GetGameEndFlag() || pi.Game.GetPhase() != domain.PopeJoanPhasePlay {
			return
		}
		idx := pi.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == popeJoanHumanIdx {
			return
		}
		handIdx := pi.Game.PopeJoanCpuDecide(idx)
		if handIdx < 0 {
			return
		}
		if err := pi.Game.Play(idx, handIdx); err != nil {
			return
		}
	}
}

// RestorePopeJoanInteractor deserialises JSON into a PopeJoanInteractor.
func RestorePopeJoanInteractor(data []byte, cp presenter.PopeJoanPresenter) (*PopeJoanInteractor, error) {
	return restoreAndBuild[domain.PopeJoan](data, func(g *domain.PopeJoan) *PopeJoanInteractor {
		return &PopeJoanInteractor{GameBase: GameBase[interfaces.PopeJoanGame]{Game: g}, cp: cp}
	})
}
