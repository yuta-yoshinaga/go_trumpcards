//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ToepenInteractorIF トゥーペンインタラクターインタフェース
type ToepenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ToepenConfig) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// Toep 賭け点を吊り上げる
	Toep() string
	// Respond toep に追随か降参かを答える
	Respond(stay bool) string
	// Redeal 貧民の手札を捨てて配り直す
	Redeal() string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ToepenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ToepenInteractor トゥーペンインタラクタークラス
type ToepenInteractor struct {
	GameBase[interfaces.ToepenGame]
	tp presenter.ToepenPresenter
}

// NewToepenInteractor コンストラクタ
func NewToepenInteractor(t interfaces.ToepenGame, tp presenter.ToepenPresenter) *ToepenInteractor {
	mustNotNil("ToepenInteractor", map[string]any{"t": t, "tp": tp})
	return &ToepenInteractor{GameBase: GameBase[interfaces.ToepenGame]{Game: t}, tp: tp}
}

// toepenHumanIdx 人間プレイヤーの座席。
const toepenHumanIdx = 0

// toepenCpuTurnCap は CPU ループの安全弁。1 ハンドは 4 トリック × 人数を超えない
// ので、これを超えるのは停止条件が壊れているとき。
const toepenCpuTurnCap = 128

// Reset ゲーム初期化
func (ti *ToepenInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *ToepenInteractor) ResetWithConfig(cfg domain.ToepenConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play 手札の札を出す
func (ti *ToepenInteractor) Play(handIdx int) string {
	if ti.Game.GetGameEndFlag() {
		return ti.tp.Output(ti.Game, nil)
	}
	if err := ti.Game.PlayCard(toepenHumanIdx, handIdx); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Toep 賭け点を吊り上げる
func (ti *ToepenInteractor) Toep() string {
	if ti.Game.GetGameEndFlag() {
		return ti.tp.Output(ti.Game, nil)
	}
	if err := ti.Game.Toep(toepenHumanIdx); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Respond toep に追随か降参かを答える
func (ti *ToepenInteractor) Respond(stay bool) string {
	if ti.Game.GetGameEndFlag() {
		return ti.tp.Output(ti.Game, nil)
	}
	if err := ti.Game.Respond(toepenHumanIdx, stay); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Redeal 貧民の手札を捨てて配り直す
func (ti *ToepenInteractor) Redeal() string {
	if ti.Game.GetGameEndFlag() {
		return ti.tp.Output(ti.Game, nil)
	}
	if err := ti.Game.Redeal(toepenHumanIdx); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextHand 次のハンドへ進む
func (ti *ToepenInteractor) NextHand() string {
	if err := ti.Game.NextHand(); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *ToepenInteractor) GetConfig() domain.ToepenConfig { return ti.Game.GetConfig() }

// Hint ヒント取得
func (ti *ToepenInteractor) Hint() string { return ti.tp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する
func (ti *ToepenInteractor) ActionLog() string { return ti.tp.ActionLogOutput(ti.Game) }

// runCpuTurns 人間の番に戻るかハンドが終わるまで CPU を進める。
//
// 応答フェーズも CPU が対象なら CPU が答える。ここで戻ると、人間が相手の
// 追随/降参を代わりに押すことになる。
func (ti *ToepenInteractor) runCpuTurns() {
	for range toepenCpuTurnCap {
		if ti.Game.GetGameEndFlag() {
			return
		}
		phase := ti.Game.GetPhase()
		if phase == domain.ToepenPhaseHandEnd || phase == domain.ToepenPhaseGameEnd {
			return
		}
		if phase == domain.ToepenPhaseRespond {
			idx := ti.Game.GetPendingRespondent()
			if idx < 0 || idx == toepenHumanIdx {
				return
			}
			action := ti.Game.ToepenCpuDecide(idx)
			if err := ti.Game.Respond(idx, !action.Fold); err != nil {
				return
			}
			continue
		}
		idx := ti.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == toepenHumanIdx {
			return
		}
		action := ti.Game.ToepenCpuDecide(idx)
		if err := ti.Game.PlayCard(idx, action.HandIdx); err != nil {
			return
		}
	}
}

// RestoreToepenInteractor deserialises JSON into a ToepenInteractor.
func RestoreToepenInteractor(data []byte, tp presenter.ToepenPresenter) (*ToepenInteractor, error) {
	return restoreAndBuild[domain.Toepen](data, func(g *domain.Toepen) *ToepenInteractor {
		return &ToepenInteractor{GameBase: GameBase[interfaces.ToepenGame]{Game: g}, tp: tp}
	})
}
