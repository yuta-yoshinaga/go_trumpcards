//go:build !js || !wasm || extra5

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OmiInteractorIF オミインタラクターインタフェース
type OmiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.OmiConfig) string
	// CallTrump 切り札スートを指名する
	CallTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OmiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OmiInteractor オミインタラクタークラス
type OmiInteractor struct {
	GameBase[interfaces.OmiGame]
	ep presenter.OmiPresenter
}

// NewOmiInteractor コンストラクタ
func NewOmiInteractor(e interfaces.OmiGame, ep presenter.OmiPresenter) *OmiInteractor {
	mustNotNil("OmiInteractor", map[string]any{"e": e, "ep": ep})
	return &OmiInteractor{GameBase: GameBase[interfaces.OmiGame]{Game: e}, ep: ep}
}

// Reset ゲーム初期化
func (ei *OmiInteractor) Reset() string {
	ei.Game.Reset()
	ei.runCpuCalls()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *OmiInteractor) ResetWithConfig(cfg domain.OmiConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.ep, cfg, ei.Game.SetConfig, ei.Reset)
}

// CallTrump スートを指名する
func (ei *OmiInteractor) CallTrump(suit int) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerCallTrump(suit)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Play カードをプレイ
func (ei *OmiInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (ei *OmiInteractor) NextTrick() string {
	ei.Game.ResolveTrick()
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.NextTrick()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ei *OmiInteractor) NextRound() string {
	ei.Game.ScoreRound()
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.NextRound()
	ei.runCpuCalls()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// GetConfig 現在の設定を取得
func (ei *OmiInteractor) GetConfig() domain.OmiConfig {
	return ei.Game.GetConfig()
}

// Hint ヒント取得
func (ei *OmiInteractor) Hint() string {
	return ei.ep.HintOutput(ei.Game)
}

// ActionLog 棋譜を出力する
func (ei *OmiInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// runCpuCalls CallTrumpフェーズでCPUが指名者の場合に自動実行する
func (ei *OmiInteractor) runCpuCalls() {
	if ei.Game.GetGameEndFlag() {
		return
	}
	if ei.Game.GetPhase() == domain.OmiPhaseCallTrump {
		if !ei.Game.IsHumanCallTrumpTurn() {
			ei.Game.CpuCallTrump()
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (ei *OmiInteractor) runCpuTurns() {
	runCpuTurnsLoop(ei.Game, trickPhases[domain.OmiPhase]{
		play:     domain.OmiPhasePlay,
		trickEnd: domain.OmiPhaseTrickEnd,
		roundEnd: domain.OmiPhaseRoundEnd,
		gameEnd:  domain.OmiPhaseGameEnd,
	})
}

// RestoreOmiInteractor deserialises JSON into an OmiInteractor.
func RestoreOmiInteractor(data []byte, ep presenter.OmiPresenter) (*OmiInteractor, error) {
	e, err := restoreGame[domain.Omi](data)
	if err != nil {
		return nil, err
	}
	return &OmiInteractor{GameBase: GameBase[interfaces.OmiGame]{Game: e}, ep: ep}, nil
}
