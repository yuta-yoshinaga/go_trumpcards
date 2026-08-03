//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DesmocheInteractorIF デスモチェインタラクターインタフェース
type DesmocheInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.DesmocheConfig) string
	// DrawStock 山札から引く
	DrawStock() string
	// DrawDiscard 捨て札を取る
	DrawDiscard() string
	// Meld 手札の添字集合をメルドとして出す
	Meld(handIdxs []int) string
	// LayOff 手札1枚をメルドに付ける
	LayOff(handIdx, meldIdx int) string
	// Desmoche 自分の場のメルドから1枚を抜いて別のメルドへ移す
	Desmoche(fromMeldIdx, cardIdx, toMeldIdx int) string
	// Discard 手札1枚を捨てる
	Discard(handIdx int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.DesmocheConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DesmocheInteractor デスモチェインタラクタークラス
type DesmocheInteractor struct {
	GameBase[interfaces.DesmocheGame]
	cp presenter.DesmochePresenter
}

// NewDesmocheInteractor コンストラクタ
func NewDesmocheInteractor(c interfaces.DesmocheGame, cp presenter.DesmochePresenter) *DesmocheInteractor {
	mustNotNil("DesmocheInteractor", map[string]any{"c": c, "cp": cp})
	return &DesmocheInteractor{GameBase: GameBase[interfaces.DesmocheGame]{Game: c}, cp: cp}
}

// desmocheHumanIdx 人間プレイヤーの座席。
const desmocheHumanIdx = 0

// desmocheCpuTurnCap は CPU ループの安全弁。
const desmocheCpuTurnCap = 4000

// Reset ゲーム初期化
func (di *DesmocheInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.cp.Output(di.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (di *DesmocheInteractor) ResetWithConfig(cfg domain.DesmocheConfig) string {
	return resetWithValidatedConfig(di.Game, di.cp, cfg, di.Game.SetConfig, di.Reset)
}

// DrawStock 山札から引く
func (di *DesmocheInteractor) DrawStock() string { return di.act(di.Game.DrawFromStock) }

// DrawDiscard 捨て札を取る
func (di *DesmocheInteractor) DrawDiscard() string { return di.act(di.Game.DrawFromDiscard) }

// act は「席番号だけを取る操作」を共通化する。
func (di *DesmocheInteractor) act(fn func(int) error) string {
	return di.apply(func() error { return fn(desmocheHumanIdx) })
}

// Meld 手札の添字集合をメルドとして出す
func (di *DesmocheInteractor) Meld(handIdxs []int) string {
	return di.apply(func() error { return di.Game.Meld(desmocheHumanIdx, handIdxs) })
}

// LayOff 手札1枚をメルドに付ける
func (di *DesmocheInteractor) LayOff(handIdx, meldIdx int) string {
	return di.apply(func() error { return di.Game.LayOff(desmocheHumanIdx, handIdx, meldIdx) })
}

// Desmoche 自分の場のメルドから1枚を抜いて別のメルドへ移す
func (di *DesmocheInteractor) Desmoche(fromMeldIdx, cardIdx, toMeldIdx int) string {
	return di.apply(func() error {
		return di.Game.Desmoche(desmocheHumanIdx, fromMeldIdx, cardIdx, toMeldIdx)
	})
}

// Discard 手札1枚を捨てる
func (di *DesmocheInteractor) Discard(handIdx int) string {
	return di.apply(func() error { return di.Game.Discard(desmocheHumanIdx, handIdx) })
}

// apply は「決着していたら何もしない → 実行 → 失敗なら伝える → CPU を回す」を
// 共通化する。
func (di *DesmocheInteractor) apply(fn func() error) string {
	if di.Game.GetGameEndFlag() {
		return di.cp.Output(di.Game, nil)
	}
	if err := fn(); err != nil {
		return di.cp.Output(di.Game, err)
	}
	di.runCpuTurns()
	return di.cp.Output(di.Game, nil)
}

// NextRound 次のラウンドへ進む
func (di *DesmocheInteractor) NextRound() string {
	if err := di.Game.NextRound(); err != nil {
		return di.cp.Output(di.Game, err)
	}
	di.runCpuTurns()
	return di.cp.Output(di.Game, nil)
}

// GetConfig 現在の設定を取得
func (di *DesmocheInteractor) GetConfig() domain.DesmocheConfig { return di.Game.GetConfig() }

// Hint ヒント取得
func (di *DesmocheInteractor) Hint() string { return di.cp.HintOutput(di.Game) }

// ActionLog 棋譜を出力する
func (di *DesmocheInteractor) ActionLog() string { return di.cp.ActionLogOutput(di.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ラウンド終了では止める。ここで勝手に配り直すと、ポットが持ち越されたのかを
// 読む間もなく画面が変わってしまう。
func (di *DesmocheInteractor) runCpuTurns() {
	for range desmocheCpuTurnCap {
		if di.Game.GetGameEndFlag() {
			return
		}
		phase := di.Game.GetPhase()
		if phase == domain.DesmochePhaseRoundEnd || phase == domain.DesmochePhaseGameEnd {
			return
		}
		idx := di.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == desmocheHumanIdx {
			return
		}
		action := di.Game.DesmocheCpuDecide(idx)
		var err error
		switch {
		case phase == domain.DesmochePhaseDraw:
			err = di.Game.DrawFromStock(idx)
		case action.MeldIdxs != nil:
			err = di.Game.Meld(idx, action.MeldIdxs)
		case action.DiscardIdx >= 0:
			err = di.Game.Discard(idx, action.DiscardIdx)
		default:
			return
		}
		if err != nil {
			return
		}
	}
}

// RestoreDesmocheInteractor deserialises JSON into a DesmocheInteractor.
func RestoreDesmocheInteractor(data []byte, cp presenter.DesmochePresenter) (*DesmocheInteractor, error) {
	return restoreAndBuild[domain.Desmoche](data, func(g *domain.Desmoche) *DesmocheInteractor {
		return &DesmocheInteractor{GameBase: GameBase[interfaces.DesmocheGame]{Game: g}, cp: cp}
	})
}
