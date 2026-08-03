//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SjavsInteractorIF シャウスインタラクターインタフェース
type SjavsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SjavsConfig) string
	// Bid 切札スート長を申告する (0 はパス)
	Bid(length int) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SjavsConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SjavsInteractor シャウスインタラクタークラス
type SjavsInteractor struct {
	GameBase[interfaces.SjavsGame]
	cp presenter.SjavsPresenter
}

// NewSjavsInteractor コンストラクタ
func NewSjavsInteractor(c interfaces.SjavsGame, cp presenter.SjavsPresenter) *SjavsInteractor {
	mustNotNil("SjavsInteractor", map[string]any{"c": c, "cp": cp})
	return &SjavsInteractor{GameBase: GameBase[interfaces.SjavsGame]{Game: c}, cp: cp}
}

// sjavsHumanIdx 人間プレイヤーの座席。
const sjavsHumanIdx = 0

// sjavsCpuTurnCap は CPU ループの安全弁。1 ハンドはビッド 4 + 32 手番なので、
// これを超えるのは停止条件が壊れているとき。
const sjavsCpuTurnCap = 128

// Reset ゲーム初期化
func (si *SjavsInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SjavsInteractor) ResetWithConfig(cfg domain.SjavsConfig) string {
	return resetWithValidatedConfig(si.Game, si.cp, cfg, si.Game.SetConfig, si.Reset)
}

// Bid 切札スート長を申告する
func (si *SjavsInteractor) Bid(length int) string {
	if si.Game.GetGameEndFlag() {
		return si.cp.Output(si.Game, nil)
	}
	if err := si.Game.Bid(sjavsHumanIdx, length); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// Play 手札の札を出す
func (si *SjavsInteractor) Play(handIdx int) string {
	if si.Game.GetGameEndFlag() {
		return si.cp.Output(si.Game, nil)
	}
	if err := si.Game.PlayCard(sjavsHumanIdx, handIdx); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// NextHand 次のハンドへ進む
func (si *SjavsInteractor) NextHand() string {
	if err := si.Game.NextHand(); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.cp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SjavsInteractor) GetConfig() domain.SjavsConfig { return si.Game.GetConfig() }

// Hint ヒント取得
func (si *SjavsInteractor) Hint() string { return si.cp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *SjavsInteractor) ActionLog() string { return si.cp.ActionLogOutput(si.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ハンド終了では止める。次のハンドへ進むかは人間の操作で、勝手に配り直すと
// 精算結果を読む間もなく画面が変わる。
func (si *SjavsInteractor) runCpuTurns() {
	for range sjavsCpuTurnCap {
		if si.Game.GetGameEndFlag() {
			return
		}
		phase := si.Game.GetPhase()
		if phase == domain.SjavsPhaseHandEnd {
			return
		}
		idx := si.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == sjavsHumanIdx {
			return
		}
		action := si.Game.SjavsCpuDecide(idx)
		var err error
		if phase == domain.SjavsPhaseBid {
			err = si.Game.Bid(idx, action.BidLength)
		} else {
			if action.HandIdx < 0 {
				return
			}
			err = si.Game.PlayCard(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// RestoreSjavsInteractor deserialises JSON into a SjavsInteractor.
func RestoreSjavsInteractor(data []byte, cp presenter.SjavsPresenter) (*SjavsInteractor, error) {
	return restoreAndBuild[domain.Sjavs](data, func(g *domain.Sjavs) *SjavsInteractor {
		return &SjavsInteractor{GameBase: GameBase[interfaces.SjavsGame]{Game: g}, cp: cp}
	})
}
