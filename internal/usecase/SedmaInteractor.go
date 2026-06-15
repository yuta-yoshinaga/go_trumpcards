//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SedmaInteractorIF セドマのインタラクターインタフェース
type SedmaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SedmaConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SedmaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SedmaInteractor セドマのインタラクタークラス
type SedmaInteractor struct {
	GameBase[interfaces.SedmaGame]
	mp presenter.SedmaPresenter
}

// NewSedmaInteractor コンストラクタ
func NewSedmaInteractor(g interfaces.SedmaGame, mp presenter.SedmaPresenter) *SedmaInteractor {
	mustNotNil("SedmaInteractor", map[string]any{"g": g, "mp": mp})
	return &SedmaInteractor{GameBase: GameBase[interfaces.SedmaGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *SedmaInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *SedmaInteractor) ResetWithConfig(cfg domain.SedmaConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *SedmaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if mi.Game.GetPhase() == domain.SedmaPhaseTrickEnd {
		mi.Game.ResolveTrick()
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *SedmaInteractor) NextTrick() string {
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *SedmaInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *SedmaInteractor) GetConfig() domain.SedmaConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *SedmaInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *SedmaInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (mi *SedmaInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.SedmaPhase]{
		play:     domain.SedmaPhasePlay,
		trickEnd: domain.SedmaPhaseTrickEnd,
		roundEnd: domain.SedmaPhaseRoundEnd,
		gameEnd:  domain.SedmaPhaseGameEnd,
	})
}

// RestoreSedmaInteractor deserialises JSON into a SedmaInteractor.
func RestoreSedmaInteractor(data []byte, mp presenter.SedmaPresenter) (*SedmaInteractor, error) {
	return restoreAndBuild[domain.Sedma](data, func(g *domain.Sedma) *SedmaInteractor {
		return &SedmaInteractor{GameBase: GameBase[interfaces.SedmaGame]{Game: g}, mp: mp}
	})
}
