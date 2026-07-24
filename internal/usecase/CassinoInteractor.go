package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CassinoInteractorIF カシノインタラクターインタフェース
type CassinoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextRound 次のラウンド開始
	NextRound() string
	// Take 場札 / ビルドを捕獲する
	Take(handIdx int, tableIdxs []int, buildIdxs []int) string
	// Build ビルドを作成する
	Build(handIdx int, tableIdxs []int, declaredValue int) string
	// Trail 場に置くだけ (捕獲しない)
	Trail(handIdx int) string
	// Hint ヒント取得
	Hint() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.CassinoConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.CassinoConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CassinoInteractor カシノインタラクター。
type CassinoInteractor struct {
	GameBase[interfaces.CassinoGame]
	cp presenter.CassinoPresenter
}

// NewCassinoInteractor コンストラクタ。
func NewCassinoInteractor(cg interfaces.CassinoGame, cp presenter.CassinoPresenter) *CassinoInteractor {
	mustNotNil("CassinoInteractor", map[string]any{"cg": cg, "cp": cp})
	return &CassinoInteractor{
		GameBase: GameBase[interfaces.CassinoGame]{Game: cg},
		cp:       cp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)
func (ci *CassinoInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// NextRound 次のラウンド開始。
func (ci *CassinoInteractor) NextRound() string {
	if ci.Game.GetGameEndFlag() {
		return ci.cp.Output(ci.Game, nil)
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// Take 場札 / ビルド捕獲。
func (ci *CassinoInteractor) Take(handIdx int, tableIdxs []int, buildIdxs []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerTake(handIdx, tableIdxs, buildIdxs)
	if err == nil && !ci.Game.GetGameEndFlag() {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// Build ビルドを作成する。
func (ci *CassinoInteractor) Build(handIdx int, tableIdxs []int, declaredValue int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerBuild(handIdx, tableIdxs, declaredValue)
	if err == nil && !ci.Game.GetGameEndFlag() {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// Trail 場に置くだけ。
func (ci *CassinoInteractor) Trail(handIdx int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerTrail(handIdx)
	if err == nil && !ci.Game.GetGameEndFlag() {
		ci.runCpuTurns()
	}
	return ci.cp.Output(ci.Game, err)
}

// Hint ヒント取得。
func (ci *CassinoInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// GetConfig 現在の設定を返す。
func (ci *CassinoInteractor) GetConfig() domain.CassinoConfig { return ci.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *CassinoInteractor) ResetWithConfig(config domain.CassinoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, config, ci.Game.SetConfig, ci.Reset)
}

// ActionLog 棋譜を出力する。
func (ci *CassinoInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// cassinoMaxCpuIterations は runCpuTurns の防御的な反復上限。
// 通常 1 ラウンドで CPU が動くのは数十回 (高々 16 ターン × 数ラウンド) で、
// 1000 を超えるなら CpuPlay または NextRound が手番を進めていない可能性が高い。
const cassinoMaxCpuIterations = 1000

// runCpuTurns ゲームが終わるか人間の手番になるまで CPU ターンを回す。
// ラウンド境界に到達した場合は自動的に NextRound し、続行する。
// 万一進行が止まった場合に備えて反復上限 (cassinoMaxCpuIterations) を設けて
// 無限ループを防ぐ。
func (ci *CassinoInteractor) runCpuTurns() {
	for i := 0; i < cassinoMaxCpuIterations; i++ {
		if ci.Game.GetGameEndFlag() || ci.Game.IsHumanTurn() {
			return
		}
		ci.Game.CpuPlay()
		// Round 終了後に NextRound を自動実行 (ゲーム終了で無ければ)
		if !ci.Game.GetGameEndFlag() && ci.Game.GetPhase() != "playerTurn" {
			ci.Game.NextRound()
		}
	}
}

// RestoreCassinoInteractor deserialises JSON into a CassinoInteractor.
func RestoreCassinoInteractor(data []byte, cp presenter.CassinoPresenter) (*CassinoInteractor, error) {
	return restoreAndBuild[domain.Cassino](data, func(g *domain.Cassino) *CassinoInteractor {
		return &CassinoInteractor{GameBase: GameBase[interfaces.CassinoGame]{Game: g}, cp: cp}
	})
}
