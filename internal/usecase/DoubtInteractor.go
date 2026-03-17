package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoubtInteractorIF ダウトインタラクターインタフェース
type DoubtInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.DoubtConfig) string
	Play(cardIndices []int, claimedValue int) string
	ResolveDoubt(doubterIndices []int) string
	SkipDoubt() string
	GetCpuDoubters() []int
	GetConfig() domain.DoubtConfig
	ResetProfile() string
	ActionLog() string
}

// DoubtInteractor ダウトインタラクタークラス
type DoubtInteractor struct {
	d  interfaces.DoubtGame
	dp presenter.DoubtPresenter
}

// NewDoubtInteractor コンストラクタ
func NewDoubtInteractor(d interfaces.DoubtGame, dp presenter.DoubtPresenter) *DoubtInteractor {
	mustNotNil("DoubtInteractor", map[string]any{"d": d, "dp": dp})
	return &DoubtInteractor{d: d, dp: dp}
}

// Reset ゲーム初期化
func (di *DoubtInteractor) Reset() string {
	di.d.Reset()
	return di.dp.Output(di.d, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (di *DoubtInteractor) ResetWithConfig(cfg domain.DoubtConfig) string {
	di.d.SetConfig(cfg)
	di.d.Reset()
	return di.dp.Output(di.d, nil)
}

// Play 人間プレイヤーがカードを出す
func (di *DoubtInteractor) Play(cardIndices []int, claimedValue int) string {
	if out, blocked := guardNotPlayable(di.d, di.dp); blocked {
		return out
	}
	err := di.d.PlayerPlay(cardIndices, claimedValue)
	return di.dp.Output(di.d, err)
}

// ResolveDoubt ダウト解決 (ダウトした人が正しかったか判定)
func (di *DoubtInteractor) ResolveDoubt(doubterIndices []int) string {
	di.d.ResolveDoubt(doubterIndices)
	di.runCpuTurns()
	return di.dp.Output(di.d, nil)
}

// SkipDoubt ダウトをスキップ (誰もダウトしなかった)
func (di *DoubtInteractor) SkipDoubt() string {
	di.d.SkipDoubt()
	di.runCpuTurns()
	return di.dp.Output(di.d, nil)
}

// GetCpuDoubters CPUダウターインデックスリスト取得
func (di *DoubtInteractor) GetCpuDoubters() []int {
	return di.d.GetCpuDoubters()
}

// GetConfig 現在の設定を取得
func (di *DoubtInteractor) GetConfig() domain.DoubtConfig {
	return di.d.GetConfig()
}

// ResetProfile メタAIプロファイルをリセット
func (di *DoubtInteractor) ResetProfile() string {
	di.d.ResetProfile()
	return di.dp.Output(di.d, nil)
}

// ActionLog 棋譜を出力する
func (di *DoubtInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.d)
}

// runCpuTurns ゲームが終わるか人間の手番またはダウトフェーズになるまでCPUターンを実行
func (di *DoubtInteractor) runCpuTurns() {
	for !di.d.GetGameEndFlag() {
		if di.d.GetPhase() == domain.DoubtPhaseDoubt {
			break
		}
		if di.d.IsHumanTurn() {
			break
		}
		di.d.CpuPlay()
	}
}
