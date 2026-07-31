//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LobaInteractorIF ロバインタラクターインタフェース
type LobaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.LobaConfig) string
	// DrawStock 山札から引く
	DrawStock() string
	// DrawDiscard 捨て札を取る
	DrawDiscard() string
	// Meld 手札の添字集合をメルドとして出す
	Meld(handIdxs []int) string
	// LayOff 手札1枚をメルドに付ける
	LayOff(handIdx, meldIdx int) string
	// Discard 手札1枚を捨てる
	Discard(handIdx int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.LobaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LobaInteractor ロバインタラクタークラス
type LobaInteractor struct {
	GameBase[interfaces.LobaGame]
	cp presenter.LobaPresenter
}

// NewLobaInteractor コンストラクタ
func NewLobaInteractor(c interfaces.LobaGame, cp presenter.LobaPresenter) *LobaInteractor {
	mustNotNil("LobaInteractor", map[string]any{"c": c, "cp": cp})
	return &LobaInteractor{GameBase: GameBase[interfaces.LobaGame]{Game: c}, cp: cp}
}

// lobaHumanIdx 人間プレイヤーの座席。
const lobaHumanIdx = 0

// lobaCpuTurnCap は CPU ループの安全弁。1 ラウンドは引く/捨てるの繰り返しで
// 長引くが、これを超えるのは停止条件が壊れているとき。
const lobaCpuTurnCap = 4000

// Reset ゲーム初期化
func (li *LobaInteractor) Reset() string {
	li.Game.Reset()
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (li *LobaInteractor) ResetWithConfig(cfg domain.LobaConfig) string {
	return resetWithValidatedConfig(li.Game, li.cp, cfg, li.Game.SetConfig, li.Reset)
}

// DrawStock 山札から引く
func (li *LobaInteractor) DrawStock() string { return li.act(li.Game.DrawFromStock) }

// DrawDiscard 捨て札を取る
func (li *LobaInteractor) DrawDiscard() string { return li.act(li.Game.DrawFromDiscard) }

// act は「席番号だけを取る操作」を共通化する。
func (li *LobaInteractor) act(fn func(int) error) string {
	if li.Game.GetGameEndFlag() {
		return li.cp.Output(li.Game, nil)
	}
	if err := fn(lobaHumanIdx); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// Meld 手札の添字集合をメルドとして出す
func (li *LobaInteractor) Meld(handIdxs []int) string {
	if li.Game.GetGameEndFlag() {
		return li.cp.Output(li.Game, nil)
	}
	if err := li.Game.Meld(lobaHumanIdx, handIdxs); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// LayOff 手札1枚をメルドに付ける
func (li *LobaInteractor) LayOff(handIdx, meldIdx int) string {
	if li.Game.GetGameEndFlag() {
		return li.cp.Output(li.Game, nil)
	}
	if err := li.Game.LayOff(lobaHumanIdx, handIdx, meldIdx); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// Discard 手札1枚を捨てる
func (li *LobaInteractor) Discard(handIdx int) string {
	if li.Game.GetGameEndFlag() {
		return li.cp.Output(li.Game, nil)
	}
	if err := li.Game.Discard(lobaHumanIdx, handIdx); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// NextRound 次のラウンドへ進む
func (li *LobaInteractor) NextRound() string {
	if err := li.Game.NextRound(); err != nil {
		return li.cp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.cp.Output(li.Game, nil)
}

// GetConfig 現在の設定を取得
func (li *LobaInteractor) GetConfig() domain.LobaConfig { return li.Game.GetConfig() }

// Hint ヒント取得
func (li *LobaInteractor) Hint() string { return li.cp.HintOutput(li.Game) }

// ActionLog 棋譜を出力する
func (li *LobaInteractor) ActionLog() string { return li.cp.ActionLogOutput(li.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ラウンド終了では止める。勝手に配り直すと精算を読む間もなく画面が変わる。
func (li *LobaInteractor) runCpuTurns() {
	for range lobaCpuTurnCap {
		if li.Game.GetGameEndFlag() {
			return
		}
		phase := li.Game.GetPhase()
		if phase == domain.LobaPhaseRoundEnd || phase == domain.LobaPhaseGameEnd {
			return
		}
		idx := li.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == lobaHumanIdx {
			return
		}
		action := li.Game.LobaCpuDecide(idx)
		var err error
		switch {
		case phase == domain.LobaPhaseDraw:
			err = li.Game.DrawFromStock(idx)
		case action.MeldIdxs != nil:
			err = li.Game.Meld(idx, action.MeldIdxs)
		case action.DiscardIdx >= 0:
			err = li.Game.Discard(idx, action.DiscardIdx)
		default:
			return
		}
		if err != nil {
			return
		}
	}
}

// RestoreLobaInteractor deserialises JSON into a LobaInteractor.
func RestoreLobaInteractor(data []byte, cp presenter.LobaPresenter) (*LobaInteractor, error) {
	return restoreAndBuild[domain.Loba](data, func(g *domain.Loba) *LobaInteractor {
		return &LobaInteractor{GameBase: GameBase[interfaces.LobaGame]{Game: g}, cp: cp}
	})
}
