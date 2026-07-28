//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// JassInteractorIF ヤス(シーバー)インタラクターインタフェース
type JassInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.JassConfig) string
	// ChooseTrump 切り札スートを指名する
	ChooseTrump(suit int) string
	// Schieben パートナーへ切り札選択を委譲する
	Schieben() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.JassConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// JassInteractor ヤス(シーバー)インタラクタークラス
type JassInteractor struct {
	GameBase[interfaces.JassGame]
	jp presenter.JassPresenter
}

// NewJassInteractor コンストラクタ
func NewJassInteractor(g interfaces.JassGame, jp presenter.JassPresenter) *JassInteractor {
	mustNotNil("JassInteractor", map[string]any{"g": g, "jp": jp})
	return &JassInteractor{GameBase: GameBase[interfaces.JassGame]{Game: g}, jp: jp}
}

// Reset ゲーム初期化
func (ji *JassInteractor) Reset() string {
	ji.Game.Reset()
	ji.runCpuBids()
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ji *JassInteractor) ResetWithConfig(cfg domain.JassConfig) string {
	return resetWithValidatedConfig(ji.Game, ji.jp, cfg, ji.Game.SetConfig, ji.Reset)
}

// ChooseTrump 切り札スートを指名する
func (ji *JassInteractor) ChooseTrump(suit int) string {
	if out, blocked := guardGameEnd(ji.Game, ji.jp); blocked {
		return out
	}
	err := ji.Game.PlayerChooseTrump(suit)
	if err != nil {
		return ji.jp.Output(ji.Game, err)
	}
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// Schieben パートナーへ切り札選択を委譲する
func (ji *JassInteractor) Schieben() string {
	if out, blocked := guardGameEnd(ji.Game, ji.jp); blocked {
		return out
	}
	err := ji.Game.PlayerSchieben()
	if err != nil {
		return ji.jp.Output(ji.Game, err)
	}
	ji.runCpuBids()
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// Play カードをプレイ
func (ji *JassInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ji.Game, ji.jp); blocked {
		return out
	}
	err := ji.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ji.jp.Output(ji.Game, err)
	}
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (ji *JassInteractor) NextTrick() string {
	ji.Game.ResolveTrick()
	if out, blocked := guardGameEnd(ji.Game, ji.jp); blocked {
		return out
	}
	ji.Game.NextTrick()
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ji *JassInteractor) NextRound() string {
	ji.Game.ScoreRound()
	if out, blocked := guardGameEnd(ji.Game, ji.jp); blocked {
		return out
	}
	ji.Game.NextRound()
	ji.runCpuBids()
	ji.runCpuTurns()
	return ji.jp.Output(ji.Game, nil)
}

// GetConfig 現在の設定を取得
func (ji *JassInteractor) GetConfig() domain.JassConfig {
	return ji.Game.GetConfig()
}

// Hint ヒント取得
func (ji *JassInteractor) Hint() string {
	return ji.jp.HintOutput(ji.Game)
}

// ActionLog 棋譜を出力する
func (ji *JassInteractor) ActionLog() string {
	return ji.jp.ActionLogOutput(ji.Game)
}

// runCpuBids ビッドフェーズでCPUを自動実行する (切り札選択 / Schieben)
func (ji *JassInteractor) runCpuBids() {
	for !ji.Game.GetGameEndFlag() {
		phase := ji.Game.GetPhase()
		if phase != domain.JassPhaseBidTrump && phase != domain.JassPhaseBidPartner {
			return
		}
		if ji.Game.IsHumanBidTurn() {
			return
		}
		ji.Game.CpuBid()
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (ji *JassInteractor) runCpuTurns() {
	runCpuTurnsLoop(ji.Game, trickPhases[domain.JassPhase]{
		play:     domain.JassPhasePlay,
		trickEnd: domain.JassPhaseTrickEnd,
		roundEnd: domain.JassPhaseRoundEnd,
		gameEnd:  domain.JassPhaseGameEnd,
	})
}

// RestoreJassInteractor deserialises JSON into a JassInteractor.
func RestoreJassInteractor(data []byte, jp presenter.JassPresenter) (*JassInteractor, error) {
	g, err := restoreGame[domain.Jass](data)
	if err != nil {
		return nil, err
	}
	return &JassInteractor{GameBase: GameBase[interfaces.JassGame]{Game: g}, jp: jp}, nil
}
