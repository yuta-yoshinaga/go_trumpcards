//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpoilFiveInteractorIF スポイル・ファイブのインタラクターインタフェース
type SpoilFiveInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SpoilFiveConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SpoilFiveConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpoilFiveInteractor スポイル・ファイブのインタラクタークラス
type SpoilFiveInteractor struct {
	GameBase[interfaces.SpoilFiveGame]
	mp presenter.SpoilFivePresenter
}

// NewSpoilFiveInteractor コンストラクタ
func NewSpoilFiveInteractor(g interfaces.SpoilFiveGame, mp presenter.SpoilFivePresenter) *SpoilFiveInteractor {
	mustNotNil("SpoilFiveInteractor", map[string]any{"g": g, "mp": mp})
	return &SpoilFiveInteractor{GameBase: GameBase[interfaces.SpoilFiveGame]{Game: g}, mp: mp}
}

// Reset ゲーム初期化
func (mi *SpoilFiveInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *SpoilFiveInteractor) ResetWithConfig(cfg domain.SpoilFiveConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *SpoilFiveInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	err := mi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if mi.Game.GetPhase() == domain.SpoilFivePhaseTrickEnd {
		mi.Game.ResolveTrick()
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (mi *SpoilFiveInteractor) NextTrick() string {
	mi.Game.NextTrick()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (mi *SpoilFiveInteractor) NextRound() string {
	mi.Game.ScoreRound()
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *SpoilFiveInteractor) GetConfig() domain.SpoilFiveConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *SpoilFiveInteractor) Hint() string {
	return mi.mp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *SpoilFiveInteractor) ActionLog() string {
	return mi.mp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (mi *SpoilFiveInteractor) runCpuTurns() {
	runCpuTurnsLoop(mi.Game, trickPhases[domain.SpoilFivePhase]{
		play:     domain.SpoilFivePhasePlay,
		trickEnd: domain.SpoilFivePhaseTrickEnd,
		roundEnd: domain.SpoilFivePhaseRoundEnd,
		gameEnd:  domain.SpoilFivePhaseGameEnd,
	})
}

// RestoreSpoilFiveInteractor deserialises JSON into a SpoilFiveInteractor.
func RestoreSpoilFiveInteractor(data []byte, mp presenter.SpoilFivePresenter) (*SpoilFiveInteractor, error) {
	return restoreAndBuild[domain.SpoilFive](data, func(g *domain.SpoilFive) *SpoilFiveInteractor {
		return &SpoilFiveInteractor{GameBase: GameBase[interfaces.SpoilFiveGame]{Game: g}, mp: mp}
	})
}
