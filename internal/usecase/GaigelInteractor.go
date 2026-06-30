//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GaigelInteractorIF ガイゲルインタラクターインタフェース
type GaigelInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GaigelConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// DeclareMarriage マリアージュを宣言する
	DeclareMarriage(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GaigelConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GaigelInteractor ガイゲルインタラクタークラス
type GaigelInteractor struct {
	GameBase[interfaces.GaigelGame]
	gp presenter.GaigelPresenter
}

// NewGaigelInteractor コンストラクタ
func NewGaigelInteractor(g interfaces.GaigelGame, gp presenter.GaigelPresenter) *GaigelInteractor {
	mustNotNil("GaigelInteractor", map[string]any{"g": g, "gp": gp})
	return &GaigelInteractor{GameBase: GameBase[interfaces.GaigelGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (gi *GaigelInteractor) Reset() string {
	gi.Game.Reset()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (gi *GaigelInteractor) ResetWithConfig(cfg domain.GaigelConfig) string {
	return resetWithValidatedConfig(gi.Game, gi.gp, cfg, gi.Game.SetConfig, gi.Reset)
}

// Play カードをプレイ
func (gi *GaigelInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// DeclareMarriage マリアージュを宣言する
func (gi *GaigelInteractor) DeclareMarriage(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	err := gi.Game.PlayerDeclareMarriage(cardIndex)
	if err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (gi *GaigelInteractor) NextTrick() string {
	gi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.NextTrick()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (gi *GaigelInteractor) NextRound() string {
	gi.Game.ScoreRound()
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.NextRound()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// GetConfig 現在の設定を取得
func (gi *GaigelInteractor) GetConfig() domain.GaigelConfig {
	return gi.Game.GetConfig()
}

// Hint ヒント取得
func (gi *GaigelInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *GaigelInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (gi *GaigelInteractor) runCpuTurns() {
	runCpuTurnsLoop(gi.Game, trickPhases[domain.GaigelPhase]{
		play:     domain.GaigelPhasePlay,
		trickEnd: domain.GaigelPhaseTrickEnd,
		roundEnd: domain.GaigelPhaseRoundEnd,
		gameEnd:  domain.GaigelPhaseGameEnd,
	})
}

// RestoreGaigelInteractor deserialises JSON into a GaigelInteractor.
func RestoreGaigelInteractor(data []byte, gp presenter.GaigelPresenter) (*GaigelInteractor, error) {
	g, err := restoreGame[domain.Gaigel](data)
	if err != nil {
		return nil, err
	}
	return &GaigelInteractor{GameBase: GameBase[interfaces.GaigelGame]{Game: g}, gp: gp}, nil
}
