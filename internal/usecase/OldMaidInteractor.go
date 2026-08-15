package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OldMaidInteractorIF ババ抜きインタラクターインタフェース
type OldMaidInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	Reset(config domain.OldMaidConfig, profileData []byte) string
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
	GameBase[interfaces.OldMaidGame]
	omp presenter.OldMaidPresenter
}

// NewOldMaidInteractor コンストラクタ
func NewOldMaidInteractor(om interfaces.OldMaidGame, omp presenter.OldMaidPresenter) *OldMaidInteractor {
	mustNotNil("OldMaidInteractor", map[string]any{"om": om, "omp": omp})
	return &OldMaidInteractor{
		GameBase: GameBase[interfaces.OldMaidGame]{Game: om},
		omp:      omp,
	}
}

// GetConfig 現在の設定を返す
func (oi *OldMaidInteractor) GetConfig() domain.OldMaidConfig {
	return oi.Game.GetConfig()
}

// Reset ゲーム初期化
func (oi *OldMaidInteractor) Reset(config domain.OldMaidConfig, profileData []byte) string {
	if err := config.Validate(); err != nil {
		return oi.omp.Output(oi.Game, err)
	}
	oi.Game.SetConfig(config)
	oi.Game.Reset()
	if len(profileData) > 0 {
		_ = oi.Game.ImportProfile(profileData)
	}
	oi.runCpuTurns()
	oi.Game.ArrangeTargetForHumanDraw()
	return oi.omp.Output(oi.Game, nil)
}

// Draw 人間プレイヤーがカードを引く
// cardIdx: 引くカードのインデックス。-1 の場合はランダム選択。
func (oi *OldMaidInteractor) Draw(cardIdx int) string {
	if out, blocked := guardNotPlayable(oi.Game, oi.omp); blocked {
		return out
	}
	err := oi.Game.PlayerDraw(cardIdx)
	if err == nil && !oi.Game.GetGameEndFlag() {
		oi.runCpuTurns()
		oi.Game.ArrangeTargetForHumanDraw()
	}
	return oi.omp.Output(oi.Game, err)
}

// Shuffle 人間プレイヤーの手札をシャッフルする
func (oi *OldMaidInteractor) Shuffle() string {
	return execAndPresent(oi.Game, oi.omp, oi.Game.ShuffleHumanHand)
}

// Reorder 人間プレイヤーの手札を並び替える
func (oi *OldMaidInteractor) Reorder(indices []int) string {
	return execAndPresent(oi.Game, oi.omp, func() error { return oi.Game.ReorderHumanHand(indices) })
}

// ResetProfile メタAIプロファイルをリセット
func (oi *OldMaidInteractor) ResetProfile() string {
	return runAndPresent(oi.Game, oi.omp, oi.Game.ResetProfile)
}

// ActionLog 棋譜を出力する
func (oi *OldMaidInteractor) ActionLog() string {
	return oi.omp.ActionLogOutput(oi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (oi *OldMaidInteractor) runCpuTurns() {
	runCpuTurnsCapped(oi.Game, func() { _ = oi.Game.CpuDraw() })
}

// RestoreOldMaidInteractor deserialises JSON into an OldMaidInteractor.
func RestoreOldMaidInteractor(data []byte, omp presenter.OldMaidPresenter) (*OldMaidInteractor, error) {
	om, err := restoreGame[domain.OldMaid](data)
	if err != nil {
		return nil, err
	}
	return &OldMaidInteractor{GameBase: GameBase[interfaces.OldMaidGame]{Game: om}, omp: omp}, nil
}
