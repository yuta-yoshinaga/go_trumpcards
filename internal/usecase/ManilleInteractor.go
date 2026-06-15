//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ManilleInteractorIF マニーユのインタラクターインタフェース
type ManilleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ManilleConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ManilleConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ManilleInteractor マニーユのインタラクタークラス
type ManilleInteractor struct {
	GameBase[interfaces.ManilleGame]
	mp presenter.ManillePresenter
}

// NewManilleInteractor コンストラクタ
func NewManilleInteractor(g interfaces.ManilleGame, mp presenter.ManillePresenter) *ManilleInteractor {
	mustNotNil("ManilleInteractor", map[string]any{"g": g, "mp": mp})
	return &ManilleInteractor{GameBase: GameBase[interfaces.ManilleGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *ManilleInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *ManilleInteractor) ResetWithConfig(cfg domain.ManilleConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *ManilleInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if mi.Game.GetPhase() == domain.ManillePhaseTrickEnd {
		mi.Game.ResolveTrick()
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *ManilleInteractor) NextTrick() string {
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *ManilleInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *ManilleInteractor) GetConfig() domain.ManilleConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *ManilleInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *ManilleInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (mi *ManilleInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.ManillePhase]{
		play:     domain.ManillePhasePlay,
		trickEnd: domain.ManillePhaseTrickEnd,
		roundEnd: domain.ManillePhaseRoundEnd,
		gameEnd:  domain.ManillePhaseGameEnd,
	})
}

// RestoreManilleInteractor deserialises JSON into a ManilleInteractor.
func RestoreManilleInteractor(data []byte, mp presenter.ManillePresenter) (*ManilleInteractor, error) {
	return restoreAndBuild[domain.Manille](data, func(g *domain.Manille) *ManilleInteractor {
		return &ManilleInteractor{GameBase: GameBase[interfaces.ManilleGame]{Game: g}, mp: mp}
	})
}
