package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PigsTailInteractorIF ぶたのしっぽインタラクターインタフェース
type PigsTailInteractorIF interface {
	// Reset ゲーム初期化
	Reset(config domain.PigsTailConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.PigsTailConfig
	// Action 人間プレイヤーのアクション (山札から1枚引く)
	Action(actionIdx int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PigsTailInteractor ぶたのしっぽインタラクタークラス
type PigsTailInteractor struct {
	pt  interfaces.PigsTailGame
	ptp presenter.PigsTailPresenter
}

// NewPigsTailInteractor コンストラクタ
func NewPigsTailInteractor(pt interfaces.PigsTailGame, ptp presenter.PigsTailPresenter) *PigsTailInteractor {
	mustNotNil("PigsTailInteractor", map[string]any{"pt": pt, "ptp": ptp})
	return &PigsTailInteractor{
		pt:  pt,
		ptp: ptp,
	}
}

// GetConfig 現在の設定を返す
func (pi *PigsTailInteractor) GetConfig() domain.PigsTailConfig {
	return pi.pt.GetConfig()
}

// Reset ゲーム初期化
func (pi *PigsTailInteractor) Reset(config domain.PigsTailConfig) string {
	if err := config.Validate(); err != nil {
		return pi.ptp.Output(pi.pt, err)
	}
	pi.pt.SetConfig(config)
	pi.pt.Reset()
	pi.runCpuTurns()
	return pi.ptp.Output(pi.pt, nil)
}

// Action 人間プレイヤーのアクション (山札から1枚引く)
func (pi *PigsTailInteractor) Action(actionIdx int) string {
	if out, blocked := guardNotPlayable(pi.pt, pi.ptp); blocked {
		return out
	}
	err := pi.pt.PlayerAction(actionIdx)
	if err == nil && !pi.pt.GetGameEndFlag() {
		pi.runCpuTurns()
	}
	return pi.ptp.Output(pi.pt, err)
}

// ActionLog 棋譜を出力する
func (pi *PigsTailInteractor) ActionLog() string {
	return pi.ptp.ActionLogOutput(pi.pt)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (pi *PigsTailInteractor) runCpuTurns() {
	for !pi.pt.GetGameEndFlag() && !pi.pt.IsHumanTurn() {
		_ = pi.pt.CpuAction()
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (pi *PigsTailInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(pi.pt)
}

// RestorePigsTailInteractor deserialises JSON into a PigsTailInteractor.
func RestorePigsTailInteractor(data []byte, ptp presenter.PigsTailPresenter) (*PigsTailInteractor, error) {
	pt, err := restoreGame[domain.PigsTail](data)
	if err != nil {
		return nil, err
	}
	return &PigsTailInteractor{pt: pt, ptp: ptp}, nil
}
