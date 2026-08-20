package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoubtInteractorIF ダウトインタラクターインタフェース
type DoubtInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.DoubtConfig, profileData []byte) string
	// Play 人間プレイヤーがカードを出す
	Play(cardIndices []int, claimedValue int, humanPlayMs int) string
	// ResolveDoubt ダウト解決
	ResolveDoubt(doubterIndices []int) string
	// SkipDoubt ダウトをスキップ
	SkipDoubt() string
	// GetCpuDoubters CPUダウターインデックスリスト取得
	GetCpuDoubters() []int
	// GetConfig 現在の設定を取得
	GetConfig() domain.DoubtConfig
	// ResetProfile メタAIプロファイルをリセット
	ResetProfile() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DoubtInteractor ダウトインタラクタークラス
type DoubtInteractor struct {
	GameBase[interfaces.DoubtGame]
	dp presenter.DoubtPresenter
}

// NewDoubtInteractor コンストラクタ
func NewDoubtInteractor(d interfaces.DoubtGame, dp presenter.DoubtPresenter) *DoubtInteractor {
	mustNotNil("DoubtInteractor", map[string]any{"d": d, "dp": dp})
	return &DoubtInteractor{GameBase: GameBase[interfaces.DoubtGame]{Game: d}, dp: dp}
}

// Reset ゲーム初期化
func (di *DoubtInteractor) Reset() string {
	return runAndPresent(di.Game, di.dp, di.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (di *DoubtInteractor) ResetWithConfig(cfg domain.DoubtConfig, profileData []byte) string {
	di.Game.SetConfig(cfg)
	di.Game.Reset()
	if len(profileData) > 0 {
		_ = di.Game.ImportProfile(profileData)
	}
	return di.dp.Output(di.Game, nil)
}

// Play 人間プレイヤーがカードを出す
func (di *DoubtInteractor) Play(cardIndices []int, claimedValue int, humanPlayMs int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerPlay(cardIndices, claimedValue, humanPlayMs)
	return di.dp.Output(di.Game, err)
}

// ResolveDoubt ダウト解決 (ダウトした人が正しかったか判定)
func (di *DoubtInteractor) ResolveDoubt(doubterIndices []int) string {
	di.Game.ResolveDoubt(doubterIndices)
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// SkipDoubt ダウトをスキップ (誰もダウトしなかった)
func (di *DoubtInteractor) SkipDoubt() string {
	di.Game.SkipDoubt()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// GetCpuDoubters CPUダウターインデックスリスト取得
func (di *DoubtInteractor) GetCpuDoubters() []int {
	return di.Game.GetCpuDoubters()
}

// GetConfig 現在の設定を取得
func (di *DoubtInteractor) GetConfig() domain.DoubtConfig {
	return di.Game.GetConfig()
}

// ResetProfile メタAIプロファイルをリセット
func (di *DoubtInteractor) ResetProfile() string {
	return runAndPresent(di.Game, di.dp, di.Game.ResetProfile)
}

// ActionLog 棋譜を出力する
func (di *DoubtInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはダウトフェーズになるまでCPUターンを実行
func (di *DoubtInteractor) runCpuTurns() {
	runCpuTurnsUntil(di.Game, func() bool {
		return di.Game.GetPhase() == domain.DoubtPhaseDoubt || di.Game.IsHumanTurn()
	}, di.Game.CpuPlay)
}

// RestoreDoubtInteractor deserialises JSON into a DoubtInteractor.
func RestoreDoubtInteractor(data []byte, dp presenter.DoubtPresenter) (*DoubtInteractor, error) {
	return restoreAndBuild[domain.Doubt](data, func(g *domain.Doubt) *DoubtInteractor {
		return &DoubtInteractor{GameBase: GameBase[interfaces.DoubtGame]{Game: g}, dp: dp}
	})
}
