package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OldMaidInteractorIF ババ抜きインタラクターインタフェース
type OldMaidInteractorIF interface {
	// Reset ゲーム初期化
	Reset(config domain.OldMaidConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.OldMaidConfig
	// Draw 人間プレイヤーがカードを引く
	Draw(cardIdx int) string
	// Shuffle 人間プレイヤーの手札をシャッフルする
	Shuffle() string
	// Reorder 人間プレイヤーの手札を並び替える
	Reorder(indices []int) string
	// ResetProfile メタAIプロファイルをリセット
	ResetProfile() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OldMaidInteractor ババ抜きインタラクタークラス
type OldMaidInteractor struct {
	om  interfaces.OldMaidGame
	omp presenter.OldMaidPresenter
}

// NewOldMaidInteractor コンストラクタ
func NewOldMaidInteractor(om interfaces.OldMaidGame, omp presenter.OldMaidPresenter) *OldMaidInteractor {
	mustNotNil("OldMaidInteractor", map[string]any{"om": om, "omp": omp})
	return &OldMaidInteractor{
		om:  om,
		omp: omp,
	}
}

// GetConfig 現在の設定を返す
func (oi *OldMaidInteractor) GetConfig() domain.OldMaidConfig {
	return oi.om.GetConfig()
}

// Reset ゲーム初期化
func (oi *OldMaidInteractor) Reset(config domain.OldMaidConfig) string {
	if err := config.Validate(); err != nil {
		return oi.omp.Output(oi.om, err)
	}
	oi.om.SetConfig(config)
	oi.om.Reset()
	oi.runCpuTurns()
	oi.om.ArrangeTargetForHumanDraw()
	return oi.omp.Output(oi.om, nil)
}

// Draw 人間プレイヤーがカードを引く
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (oi *OldMaidInteractor) Draw(cardIdx int) string {
	if out, blocked := guardNotPlayable(oi.om, oi.omp); blocked {
		return out
	}
	err := oi.om.PlayerDraw(cardIdx)
	if err == nil && !oi.om.GetGameEndFlag() {
		oi.runCpuTurns()
		oi.om.ArrangeTargetForHumanDraw()
	}
	return oi.omp.Output(oi.om, err)
}

// Shuffle 人間プレイヤーの手札をシャッフルする
func (oi *OldMaidInteractor) Shuffle() string {
	return execAndPresent(oi.om, oi.omp, oi.om.ShuffleHumanHand)
}

// Reorder 人間プレイヤーの手札を並び替える
func (oi *OldMaidInteractor) Reorder(indices []int) string {
	return execAndPresent(oi.om, oi.omp, func() error { return oi.om.ReorderHumanHand(indices) })
}

// ResetProfile メタAIプロファイルをリセット
func (oi *OldMaidInteractor) ResetProfile() string {
	return runAndPresent(oi.om, oi.omp, oi.om.ResetProfile)
}

// ActionLog 棋譜を出力する
func (oi *OldMaidInteractor) ActionLog() string {
	return oi.omp.ActionLogOutput(oi.om)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (oi *OldMaidInteractor) runCpuTurns() {
	for !oi.om.GetGameEndFlag() && !oi.om.IsHumanTurn() {
		_ = oi.om.CpuDraw()
	}
}
