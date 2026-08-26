//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrappolaInteractorIF トラッポラのインタラクターインタフェース
type TrappolaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TrappolaConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TrappolaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrappolaInteractor トラッポラのインタラクタークラス
type TrappolaInteractor struct {
	GameBase[interfaces.TrappolaGame]
	tp presenter.TrappolaPresenter
}

// NewTrappolaInteractor コンストラクタ
func NewTrappolaInteractor(g interfaces.TrappolaGame, tp presenter.TrappolaPresenter) *TrappolaInteractor {
	mustNotNil("TrappolaInteractor", map[string]any{"g": g, "tp": tp})
	return &TrappolaInteractor{GameBase: GameBase[interfaces.TrappolaGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TrappolaInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TrappolaInteractor) ResetWithConfig(cfg domain.TrappolaConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *TrappolaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.TrappolaPhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TrappolaInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TrappolaInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TrappolaInteractor) GetConfig() domain.TrappolaConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TrappolaInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TrappolaInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (ti *TrappolaInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TrappolaPhase]{
		play:     domain.TrappolaPhasePlay,
		trickEnd: domain.TrappolaPhaseTrickEnd,
		roundEnd: domain.TrappolaPhaseRoundEnd,
		gameEnd:  domain.TrappolaPhaseGameEnd,
	})
}

// RestoreTrappolaInteractor deserialises JSON into a TrappolaInteractor.
func RestoreTrappolaInteractor(data []byte, tp presenter.TrappolaPresenter) (*TrappolaInteractor, error) {
	return restoreAndBuild[domain.Trappola](data, func(g *domain.Trappola) *TrappolaInteractor {
		return &TrappolaInteractor{GameBase: GameBase[interfaces.TrappolaGame]{Game: g}, tp: tp}
	})
}
