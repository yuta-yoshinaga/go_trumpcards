//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrexInteractorIF トリックスインタラクターインタフェース
type TrexInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TrexConfig) string
	// Choose 王が契約を選ぶ
	Choose(contract int) string
	// Play 手札の札を出す
	Play(handIdx int) string
	// Pass ドミノでパスする
	Pass() string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TrexConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrexInteractor トリックスインタラクタークラス
type TrexInteractor struct {
	GameBase[interfaces.TrexGame]
	cp presenter.TrexPresenter
}

// NewTrexInteractor コンストラクタ
func NewTrexInteractor(c interfaces.TrexGame, cp presenter.TrexPresenter) *TrexInteractor {
	mustNotNil("TrexInteractor", map[string]any{"c": c, "cp": cp})
	return &TrexInteractor{GameBase: GameBase[interfaces.TrexGame]{Game: c}, cp: cp}
}

// trexHumanIdx 人間プレイヤーの座席。
const trexHumanIdx = 0

// trexCpuTurnCap は CPU ループの安全弁。1 ディールは 52 手番を超えないので、
// これを超えるのは停止条件が壊れているとき。
const trexCpuTurnCap = 256

// Reset ゲーム初期化
func (ti *TrexInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.cp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TrexInteractor) ResetWithConfig(cfg domain.TrexConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.cp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Choose 王が契約を選ぶ
func (ti *TrexInteractor) Choose(contract int) string {
	if ti.Game.GetGameEndFlag() {
		return ti.cp.Output(ti.Game, nil)
	}
	if err := ti.Game.ChooseContract(trexHumanIdx, domain.TrexContract(contract)); err != nil {
		return ti.cp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.cp.Output(ti.Game, nil)
}

// Play 手札の札を出す
func (ti *TrexInteractor) Play(handIdx int) string {
	if ti.Game.GetGameEndFlag() {
		return ti.cp.Output(ti.Game, nil)
	}
	if err := ti.Game.PlayCard(trexHumanIdx, handIdx); err != nil {
		return ti.cp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.cp.Output(ti.Game, nil)
}

// Pass ドミノでパスする
func (ti *TrexInteractor) Pass() string {
	if ti.Game.GetGameEndFlag() {
		return ti.cp.Output(ti.Game, nil)
	}
	if err := ti.Game.Pass(trexHumanIdx); err != nil {
		return ti.cp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.cp.Output(ti.Game, nil)
}

// NextDeal 次のディールへ進む
func (ti *TrexInteractor) NextDeal() string {
	if err := ti.Game.NextDeal(); err != nil {
		return ti.cp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.cp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TrexInteractor) GetConfig() domain.TrexConfig { return ti.Game.GetConfig() }

// Hint ヒント取得
func (ti *TrexInteractor) Hint() string { return ti.cp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する
func (ti *TrexInteractor) ActionLog() string { return ti.cp.ActionLogOutput(ti.Game) }

// runCpuTurns 人間の手番に戻るか終局するまで CPU を進める。
//
// ディール終了では止める。勝手に配り直すと精算を読む間もなく画面が変わる。
// **契約選択は王が CPU のときだけ CPU が行う** —— 人間が王のときにここで
// 選んでしまうと、issue の肝である「王が契約を選ぶ」が消える。
func (ti *TrexInteractor) runCpuTurns() {
	for range trexCpuTurnCap {
		if ti.Game.GetGameEndFlag() {
			return
		}
		switch ti.Game.GetPhase() {
		case domain.TrexPhaseDealEnd, domain.TrexPhaseGameEnd:
			return
		case domain.TrexPhaseChoose:
			king := ti.Game.GetKingIdx()
			if king == trexHumanIdx {
				return
			}
			action := ti.Game.TrexCpuDecide(king)
			if err := ti.Game.ChooseContract(king, action.Contract); err != nil {
				return
			}
		case domain.TrexPhasePlay:
			idx := ti.Game.GetCurrentPlayerIdx()
			if idx < 0 || idx == trexHumanIdx {
				return
			}
			action := ti.Game.TrexCpuDecide(idx)
			var err error
			if action.Pass {
				err = ti.Game.Pass(idx)
			} else if action.HandIdx < 0 {
				return
			} else {
				err = ti.Game.PlayCard(idx, action.HandIdx)
			}
			if err != nil {
				return
			}
		}
	}
}

// RestoreTrexInteractor deserialises JSON into a TrexInteractor.
func RestoreTrexInteractor(data []byte, cp presenter.TrexPresenter) (*TrexInteractor, error) {
	return restoreAndBuild[domain.Trex](data, func(g *domain.Trex) *TrexInteractor {
		return &TrexInteractor{GameBase: GameBase[interfaces.TrexGame]{Game: g}, cp: cp}
	})
}
