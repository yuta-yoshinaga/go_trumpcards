//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BuraInteractorIF ブラインタラクターインタフェース
type BuraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BuraConfig) string
	// Play 手札の indices を出す
	Play(indices []int) string
	// Claim 31点到達を宣言する
	Claim() string
	// Declare 手札の即勝ち役を宣言する
	Declare() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BuraConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BuraInteractor ブラインタラクタークラス
type BuraInteractor struct {
	GameBase[interfaces.BuraGame]
	bp presenter.BuraPresenter
}

// NewBuraInteractor コンストラクタ
func NewBuraInteractor(b interfaces.BuraGame, bp presenter.BuraPresenter) *BuraInteractor {
	mustNotNil("BuraInteractor", map[string]any{"b": b, "bp": bp})
	return &BuraInteractor{GameBase: GameBase[interfaces.BuraGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BuraInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BuraInteractor) ResetWithConfig(cfg domain.BuraConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play 手札の indices を出す
func (bi *BuraInteractor) Play(indices []int) string {
	if bi.Game.GetGameEndFlag() {
		return bi.bp.Output(bi.Game, nil)
	}
	if err := bi.Game.PlayCards(buraHumanIdx, indices); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Claim 31点到達を宣言する。足りていなければ相手の勝ちとなる。
func (bi *BuraInteractor) Claim() string {
	if bi.Game.GetGameEndFlag() {
		return bi.bp.Output(bi.Game, nil)
	}
	if err := bi.Game.Claim(buraHumanIdx); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	return bi.bp.Output(bi.Game, nil)
}

// Declare 手札の即勝ち役を宣言する。役がなければエラーを返して続行する。
func (bi *BuraInteractor) Declare() string {
	if bi.Game.GetGameEndFlag() {
		return bi.bp.Output(bi.Game, nil)
	}
	if err := bi.Game.DeclareCombination(buraHumanIdx); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BuraInteractor) GetConfig() domain.BuraConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BuraInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BuraInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// buraHumanIdx 人間プレイヤーの座席。
const buraHumanIdx = 0

// buraCpuTurnCap は CPU ループの安全弁。1 ラウンドは 36 枚を高々 1 枚ずつ
// 消費するので 36 手を超えることはない。超えたらループの停止条件が壊れて
// いるので、無限ループにするより打ち切る。
const buraCpuTurnCap = 64

// runCpuTurns 人間の手番に戻るかゲームが終わるまで CPU を進める。
func (bi *BuraInteractor) runCpuTurns() {
	for range buraCpuTurnCap {
		if bi.Game.GetGameEndFlag() {
			return
		}
		idx := bi.Game.GetCurrentPlayerIdx()
		if idx < 0 || idx == buraHumanIdx {
			return
		}
		action := bi.Game.BuraCpuDecide(idx)
		switch {
		case action.Declare:
			if err := bi.Game.DeclareCombination(idx); err != nil {
				return
			}
		case action.Claim:
			if err := bi.Game.Claim(idx); err != nil {
				return
			}
		default:
			if err := bi.Game.PlayCards(idx, action.Indices); err != nil {
				return
			}
		}
	}
}

// RestoreBuraInteractor deserialises JSON into a BuraInteractor.
func RestoreBuraInteractor(data []byte, bp presenter.BuraPresenter) (*BuraInteractor, error) {
	return restoreAndBuild[domain.Bura](data, func(g *domain.Bura) *BuraInteractor {
		return &BuraInteractor{GameBase: GameBase[interfaces.BuraGame]{Game: g}, bp: bp}
	})
}
