//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MariasInteractorIF マリアーシュのインタラクターインタフェース
type MariasInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MariasConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MariasConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MariasInteractor マリアーシュのインタラクタークラス
type MariasInteractor struct {
	GameBase[interfaces.MariasGame]
	mp presenter.MariasPresenter
}

// NewMariasInteractor コンストラクタ
func NewMariasInteractor(g interfaces.MariasGame, mp presenter.MariasPresenter) *MariasInteractor {
	mustNotNil("MariasInteractor", map[string]any{"g": g, "mp": mp})
	return &MariasInteractor{GameBase: GameBase[interfaces.MariasGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *MariasInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MariasInteractor) ResetWithConfig(cfg domain.MariasConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *MariasInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if mi.Game.GetPhase() == domain.MariasPhaseTrickEnd {
		mi.Game.ResolveTrick()
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *MariasInteractor) NextTrick() string {
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *MariasInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MariasInteractor) GetConfig() domain.MariasConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *MariasInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MariasInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (mi *MariasInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.MariasPhase]{
		play:     domain.MariasPhasePlay,
		trickEnd: domain.MariasPhaseTrickEnd,
		roundEnd: domain.MariasPhaseRoundEnd,
		gameEnd:  domain.MariasPhaseGameEnd,
	})
}

// RestoreMariasInteractor deserialises JSON into a MariasInteractor.
func RestoreMariasInteractor(data []byte, mp presenter.MariasPresenter) (*MariasInteractor, error) {
	return restoreAndBuild[domain.Marias](data, func(g *domain.Marias) *MariasInteractor {
		return &MariasInteractor{GameBase: GameBase[interfaces.MariasGame]{Game: g}, mp: mp}
	})
}
