//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NainJauneInteractorIF ル・ナン・ジョーヌインタラクターインタフェース
type NainJauneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.NainJauneConfig) string
	// Play 手札1枚を出す
	Play(handIdx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.NainJauneConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// NainJauneInteractor ル・ナン・ジョーヌインタラクタークラス
type NainJauneInteractor struct {
	GameBase[interfaces.NainJauneGame]
	cp presenter.NainJaunePresenter
}

// NewNainJauneInteractor コンストラクタ
func NewNainJauneInteractor(c interfaces.NainJauneGame, cp presenter.NainJaunePresenter) *NainJauneInteractor {
	mustNotNil("NainJauneInteractor", map[string]any{"c": c, "cp": cp})
	return &NainJauneInteractor{GameBase: GameBase[interfaces.NainJauneGame]{Game: c}, cp: cp}
}

// nainJauneHumanIdx 人間プレイヤーの座席。
const nainJauneHumanIdx = 0

// nainJauneCpuTurnCap は CPU ループの安全弁。
const nainJauneCpuTurnCap = 2000

// Reset ゲーム初期化
func (pi *NainJauneInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *NainJauneInteractor) ResetWithConfig(cfg domain.NainJauneConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.cp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Play 手札1枚を出す
func (pi *NainJauneInteractor) Play(handIdx int) string {
	if pi.Game.GetGameEndFlag() {
		return pi.cp.Output(pi.Game, nil)
	}
	if err := pi.Game.Play(nainJauneHumanIdx, handIdx); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// NextDeal 次のディールへ進む
func (pi *NainJauneInteractor) NextDeal() string {
	if err := pi.Game.NextDeal(); err != nil {
		return pi.cp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.cp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *NainJauneInteractor) GetConfig() domain.NainJauneConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *NainJauneInteractor) Hint() string { return pi.cp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *NainJauneInteractor) ActionLog() string { return pi.cp.ActionLogOutput(pi.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ディール終了では止める。勝手に配り直すと、5 区画の精算を読む間もなく画面が
// 変わってしまう。
func (pi *NainJauneInteractor) runCpuTurns() {
	for range nainJauneCpuTurnCap {
		if pi.Game.GetGameEndFlag() || pi.Game.GetPhase() != domain.NainJaunePhasePlay {
			return
		}
		idx := pi.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == nainJauneHumanIdx {
			return
		}
		handIdx := pi.Game.NainJauneCpuDecide(idx)
		if handIdx < 0 {
			return
		}
		if err := pi.Game.Play(idx, handIdx); err != nil {
			return
		}
	}
}

// RestoreNainJauneInteractor deserialises JSON into a NainJauneInteractor.
func RestoreNainJauneInteractor(data []byte, cp presenter.NainJaunePresenter) (*NainJauneInteractor, error) {
	return restoreAndBuild[domain.NainJaune](data, func(g *domain.NainJaune) *NainJauneInteractor {
		return &NainJauneInteractor{GameBase: GameBase[interfaces.NainJauneGame]{Game: g}, cp: cp}
	})
}
