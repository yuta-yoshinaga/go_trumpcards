//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HachiHachiInteractorIF は八八 (Hachi-Hachi) のインタラクターインタフェース。
type HachiHachiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.HachiHachiConfig) string
	// Play 手札を出す (fieldIdx で 2 枚一致時の捕獲対象を指定; 不要なら -1)
	Play(handIdx, fieldIdx int) string
	// NextRound 次のラウンドを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.HachiHachiConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HachiHachiInteractor は八八インタラクター。
type HachiHachiInteractor struct {
	GameBase[interfaces.HachiHachiGame]
	cp presenter.HachiHachiPresenter
}

// NewHachiHachiInteractor コンストラクタ。
func NewHachiHachiInteractor(kg interfaces.HachiHachiGame, cp presenter.HachiHachiPresenter) *HachiHachiInteractor {
	mustNotNil("HachiHachiInteractor", map[string]any{"kg": kg, "cp": cp})
	return &HachiHachiInteractor{GameBase: GameBase[interfaces.HachiHachiGame]{Game: kg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ki *HachiHachiInteractor) Reset() string {
	ki.Game.Reset()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ki *HachiHachiInteractor) ResetWithConfig(cfg domain.HachiHachiConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.cp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Play 手札を出す。
func (ki *HachiHachiInteractor) Play(handIdx, fieldIdx int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.cp); blocked {
		return out
	}
	if err := ki.Game.PlayerPlay(handIdx, fieldIdx); err != nil {
		return ki.cp.Output(ki.Game, err)
	}
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// NextRound 次のラウンドを開始する。
func (ki *HachiHachiInteractor) NextRound() string {
	ki.Game.NextRound()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ki *HachiHachiInteractor) GetConfig() domain.HachiHachiConfig { return ki.Game.GetConfig() }

// Hint ヒントを出力する。
func (ki *HachiHachiInteractor) Hint() string { return ki.cp.HintOutput(ki.Game) }

// ActionLog 棋譜を出力する。
func (ki *HachiHachiInteractor) ActionLog() string { return ki.cp.ActionLogOutput(ki.Game) }

// hachihachiMaxCpuIterations は advance の防御的な反復上限。
const hachihachiMaxCpuIterations = 1000

// advance はゲーム終了・ラウンド終了・人間の手番のいずれかに到達するまで CPU の
// プレイを回す。
func (ki *HachiHachiInteractor) advance() {
	for i := 0; i < hachihachiMaxCpuIterations; i++ {
		if ki.Game.GetGameEndFlag() {
			return
		}
		if ki.Game.GetPhase() != domain.HachiHachiPhasePlay {
			// RoundEnd / GameEnd は人間の操作待ち。
			return
		}
		if ki.Game.IsHumanTurn() {
			return
		}
		ki.Game.CpuPlay()
	}
}

// RestoreHachiHachiInteractor deserialises JSON into a HachiHachiInteractor.
func RestoreHachiHachiInteractor(data []byte, cp presenter.HachiHachiPresenter) (*HachiHachiInteractor, error) {
	return restoreAndBuild[domain.HachiHachi](data, func(g *domain.HachiHachi) *HachiHachiInteractor {
		return &HachiHachiInteractor{GameBase: GameBase[interfaces.HachiHachiGame]{Game: g}, cp: cp}
	})
}
