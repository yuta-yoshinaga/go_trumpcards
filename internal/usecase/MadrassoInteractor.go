//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MadrassoInteractorIF マドラッソのインタラクターインタフェース
type MadrassoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MadrassoConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MadrassoConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MadrassoInteractor マドラッソのインタラクタークラス
type MadrassoInteractor struct {
	GameBase[interfaces.MadrassoGame]
	tp presenter.MadrassoPresenter
}

// NewMadrassoInteractor コンストラクタ
func NewMadrassoInteractor(g interfaces.MadrassoGame, tp presenter.MadrassoPresenter) *MadrassoInteractor {
	mustNotNil("MadrassoInteractor", map[string]any{"g": g, "tp": tp})
	return &MadrassoInteractor{GameBase: GameBase[interfaces.MadrassoGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化
func (ti *MadrassoInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *MadrassoInteractor) ResetWithConfig(cfg domain.MadrassoConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *MadrassoInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.MadrassoPhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *MadrassoInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *MadrassoInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *MadrassoInteractor) GetConfig() domain.MadrassoConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *MadrassoInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *MadrassoInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (ti *MadrassoInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.MadrassoPhase]{
		play:     domain.MadrassoPhasePlay,
		trickEnd: domain.MadrassoPhaseTrickEnd,
		roundEnd: domain.MadrassoPhaseRoundEnd,
		gameEnd:  domain.MadrassoPhaseGameEnd,
	})
}

// RestoreMadrassoInteractor deserialises JSON into a MadrassoInteractor.
func RestoreMadrassoInteractor(data []byte, tp presenter.MadrassoPresenter) (*MadrassoInteractor, error) {
	return restoreAndBuild[domain.Madrasso](data, func(g *domain.Madrasso) *MadrassoInteractor {
		return &MadrassoInteractor{GameBase: GameBase[interfaces.MadrassoGame]{Game: g}, tp: tp}
	})
}
