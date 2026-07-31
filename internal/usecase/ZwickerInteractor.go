//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ZwickerInteractorIF ツヴィッカーインタラクターインタフェース
type ZwickerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ZwickerConfig) string
	// Take 手札1枚を指定の値として使い、場札とビルドを取る
	Take(handIdx, playedValue int, tableIdxs, buildIdxs []int) string
	// Build 手札1枚と場札を積んで宣言値のビルドを作る
	Build(handIdx int, tableIdxs []int, declaredValue int) string
	// Trail 手札1枚を場に置いて手番を終える
	Trail(handIdx int) string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ZwickerConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ZwickerInteractor ツヴィッカーインタラクタークラス
type ZwickerInteractor struct {
	GameBase[interfaces.ZwickerGame]
	cp presenter.ZwickerPresenter
}

// NewZwickerInteractor コンストラクタ
func NewZwickerInteractor(c interfaces.ZwickerGame, cp presenter.ZwickerPresenter) *ZwickerInteractor {
	mustNotNil("ZwickerInteractor", map[string]any{"c": c, "cp": cp})
	return &ZwickerInteractor{GameBase: GameBase[interfaces.ZwickerGame]{Game: c}, cp: cp}
}

// zwickerHumanIdx 人間プレイヤーの座席。
const zwickerHumanIdx = 0

// zwickerCpuTurnCap は CPU ループの安全弁。1 ディールは 55 枚を配り切るまで
// 続くが、これを超えるのは停止条件が壊れているとき。
const zwickerCpuTurnCap = 2000

// Reset ゲーム初期化
func (zi *ZwickerInteractor) Reset() string {
	zi.Game.Reset()
	zi.runCpuTurns()
	return zi.cp.Output(zi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (zi *ZwickerInteractor) ResetWithConfig(cfg domain.ZwickerConfig) string {
	return resetWithValidatedConfig(zi.Game, zi.cp, cfg, zi.Game.SetConfig, zi.Reset)
}

// Take 手札1枚を指定の値として使い、場札とビルドを取る
func (zi *ZwickerInteractor) Take(handIdx, playedValue int, tableIdxs, buildIdxs []int) string {
	return zi.apply(func() error {
		return zi.Game.Take(zwickerHumanIdx, handIdx, playedValue, tableIdxs, buildIdxs)
	})
}

// Build 手札1枚と場札を積んで宣言値のビルドを作る
func (zi *ZwickerInteractor) Build(handIdx int, tableIdxs []int, declaredValue int) string {
	return zi.apply(func() error {
		return zi.Game.Build(zwickerHumanIdx, handIdx, tableIdxs, declaredValue)
	})
}

// Trail 手札1枚を場に置いて手番を終える
func (zi *ZwickerInteractor) Trail(handIdx int) string {
	return zi.apply(func() error { return zi.Game.Trail(zwickerHumanIdx, handIdx) })
}

// apply は「決着していたら何もしない → 実行 → 失敗なら伝える → CPU を回す」を
// 共通化する。
func (zi *ZwickerInteractor) apply(fn func() error) string {
	if zi.Game.GetGameEndFlag() {
		return zi.cp.Output(zi.Game, nil)
	}
	if err := fn(); err != nil {
		return zi.cp.Output(zi.Game, err)
	}
	zi.runCpuTurns()
	return zi.cp.Output(zi.Game, nil)
}

// NextRound 次のディールへ進む
func (zi *ZwickerInteractor) NextRound() string {
	if err := zi.Game.NextRound(); err != nil {
		return zi.cp.Output(zi.Game, err)
	}
	zi.runCpuTurns()
	return zi.cp.Output(zi.Game, nil)
}

// GetConfig 現在の設定を取得
func (zi *ZwickerInteractor) GetConfig() domain.ZwickerConfig { return zi.Game.GetConfig() }

// Hint ヒント取得
func (zi *ZwickerInteractor) Hint() string { return zi.cp.HintOutput(zi.Game) }

// ActionLog 棋譜を出力する
func (zi *ZwickerInteractor) ActionLog() string { return zi.cp.ActionLogOutput(zi.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ディール終了では止める。勝手に配り直すと、30 点の内訳を読む間もなく画面が
// 変わってしまう。
func (zi *ZwickerInteractor) runCpuTurns() {
	for range zwickerCpuTurnCap {
		if zi.Game.GetGameEndFlag() || zi.Game.GetPhase() != domain.ZwickerPhasePlay {
			return
		}
		idx := zi.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == zwickerHumanIdx {
			return
		}
		action := zi.Game.ZwickerCpuDecide(idx)
		if action.HandIdx < 0 {
			return
		}
		var err error
		if action.Type == "take" {
			err = zi.Game.Take(idx, action.HandIdx, action.Value, action.TableIdxs, nil)
		} else {
			err = zi.Game.Trail(idx, action.HandIdx)
		}
		if err != nil {
			return
		}
	}
}

// RestoreZwickerInteractor deserialises JSON into a ZwickerInteractor.
func RestoreZwickerInteractor(data []byte, cp presenter.ZwickerPresenter) (*ZwickerInteractor, error) {
	return restoreAndBuild[domain.Zwicker](data, func(g *domain.Zwicker) *ZwickerInteractor {
		return &ZwickerInteractor{GameBase: GameBase[interfaces.ZwickerGame]{Game: g}, cp: cp}
	})
}
