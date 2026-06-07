//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TressetteInteractorIF トレセッテのインタラクターインタフェース
type TressetteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TressetteConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TressetteConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TressetteInteractor トレセッテのインタラクタークラス
type TressetteInteractor struct {
	GameBase[interfaces.TressetteGame]
	tp presenter.TressettePresenter
}

// NewTressetteInteractor コンストラクタ
func NewTressetteInteractor(g interfaces.TressetteGame, tp presenter.TressettePresenter) *TressetteInteractor {
	mustNotNil("TressetteInteractor", map[string]any{"g": g, "tp": tp})
	return &TressetteInteractor{GameBase: GameBase[interfaces.TressetteGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TressetteInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TressetteInteractor) ResetWithConfig(cfg domain.TressetteConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *TressetteInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.TressettePhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TressetteInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TressetteInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TressetteInteractor) GetConfig() domain.TressetteConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TressetteInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TressetteInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (ti *TressetteInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TressettePhase]{
		play:     domain.TressettePhasePlay,
		trickEnd: domain.TressettePhaseTrickEnd,
		roundEnd: domain.TressettePhaseRoundEnd,
		gameEnd:  domain.TressettePhaseGameEnd,
	})
}

// RestoreTressetteInteractor deserialises JSON into a TressetteInteractor.
func RestoreTressetteInteractor(data []byte, tp presenter.TressettePresenter) (*TressetteInteractor, error) {
	return restoreAndBuild[domain.Tressette](data, func(g *domain.Tressette) *TressetteInteractor {
		return &TressetteInteractor{GameBase: GameBase[interfaces.TressetteGame]{Game: g}, tp: tp}
	})
}
