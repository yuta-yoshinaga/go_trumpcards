//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KnockoutWhistInteractorIF ノックアウト・ホイストのインタラクターインタフェース
type KnockoutWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KnockoutWhistConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KnockoutWhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KnockoutWhistInteractor ノックアウト・ホイストのインタラクタークラス
type KnockoutWhistInteractor struct {
	GameBase[interfaces.KnockoutWhistGame]
	mp presenter.KnockoutWhistPresenter
}

// NewKnockoutWhistInteractor コンストラクタ
func NewKnockoutWhistInteractor(g interfaces.KnockoutWhistGame, mp presenter.KnockoutWhistPresenter) *KnockoutWhistInteractor {
	mustNotNil("KnockoutWhistInteractor", map[string]any{"g": g, "mp": mp})
	return &KnockoutWhistInteractor{GameBase: GameBase[interfaces.KnockoutWhistGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *KnockoutWhistInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *KnockoutWhistInteractor) ResetWithConfig(cfg domain.KnockoutWhistConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *KnockoutWhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if mi.Game.GetPhase() == domain.KnockoutWhistPhaseTrickEnd {
		mi.Game.ResolveTrick()
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *KnockoutWhistInteractor) NextTrick() string {
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *KnockoutWhistInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *KnockoutWhistInteractor) GetConfig() domain.KnockoutWhistConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *KnockoutWhistInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *KnockoutWhistInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (mi *KnockoutWhistInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.KnockoutWhistPhase]{
		play:     domain.KnockoutWhistPhasePlay,
		trickEnd: domain.KnockoutWhistPhaseTrickEnd,
		roundEnd: domain.KnockoutWhistPhaseRoundEnd,
		gameEnd:  domain.KnockoutWhistPhaseGameEnd,
	})
}

// RestoreKnockoutWhistInteractor deserialises JSON into a KnockoutWhistInteractor.
func RestoreKnockoutWhistInteractor(data []byte, mp presenter.KnockoutWhistPresenter) (*KnockoutWhistInteractor, error) {
	return restoreAndBuild[domain.KnockoutWhist](data, func(g *domain.KnockoutWhist) *KnockoutWhistInteractor {
		return &KnockoutWhistInteractor{GameBase: GameBase[interfaces.KnockoutWhistGame]{Game: g}, mp: mp}
	})
}
