//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TuteInteractorIF トゥーテのインタラクターインタフェース
type TuteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TuteConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// DeclareMarriage 結婚宣言をする
	DeclareMarriage(suit int) string
	// DeclareTute Tute を宣言する
	DeclareTute() string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TuteConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TuteInteractor トゥーテのインタラクタークラス
type TuteInteractor struct {
	GameBase[interfaces.TuteGame]
	tp presenter.TutePresenter
}

// NewTuteInteractor コンストラクタ
func NewTuteInteractor(g interfaces.TuteGame, tp presenter.TutePresenter) *TuteInteractor {
	mustNotNil("TuteInteractor", map[string]any{"g": g, "tp": tp})
	return &TuteInteractor{GameBase: GameBase[interfaces.TuteGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TuteInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TuteInteractor) ResetWithConfig(cfg domain.TuteConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Play カードをプレイ
func (ti *TuteInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ti.Game.GetPhase() == domain.TutePhaseTrickEnd {
		ti.Game.ResolveTrick()
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// DeclareMarriage 結婚宣言をする
func (ti *TuteInteractor) DeclareMarriage(suit int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDeclareMarriage(suit); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	return ti.tp.Output(ti.Game, nil)
}

// DeclareTute Tute を宣言する
func (ti *TuteInteractor) DeclareTute() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDeclareTute(); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TuteInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TuteInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TuteInteractor) GetConfig() domain.TuteConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TuteInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TuteInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (ti *TuteInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TutePhase]{
		play:     domain.TutePhasePlay,
		trickEnd: domain.TutePhaseTrickEnd,
		roundEnd: domain.TutePhaseRoundEnd,
		gameEnd:  domain.TutePhaseGameEnd,
	})
}

// RestoreTuteInteractor deserialises JSON into a TuteInteractor.
func RestoreTuteInteractor(data []byte, tp presenter.TutePresenter) (*TuteInteractor, error) {
	return restoreAndBuild[domain.Tute](data, func(g *domain.Tute) *TuteInteractor {
		return &TuteInteractor{GameBase: GameBase[interfaces.TuteGame]{Game: g}, tp: tp}
	})
}
