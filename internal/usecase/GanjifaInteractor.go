//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GanjifaInteractorIF ガンジファのインタラクターインタフェース
type GanjifaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GanjifaConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GanjifaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GanjifaInteractor ガンジファのインタラクタークラス
type GanjifaInteractor struct {
	GameBase[interfaces.GanjifaGame]
	sp presenter.GanjifaPresenter
}

// NewGanjifaInteractor コンストラクタ
func NewGanjifaInteractor(g interfaces.GanjifaGame, sp presenter.GanjifaPresenter) *GanjifaInteractor {
	mustNotNil("GanjifaInteractor", map[string]any{"g": g, "sp": sp})
	return &GanjifaInteractor{GameBase: GameBase[interfaces.GanjifaGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (pi *GanjifaInteractor) Reset() string {
	pi.Game.Reset()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *GanjifaInteractor) ResetWithConfig(cfg domain.GanjifaConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.sp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Play カードをプレイ
func (pi *GanjifaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerPlay(cardIndex); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if pi.Game.GetPhase() == domain.GanjifaPhaseTrickEnd {
		pi.Game.ResolveTrick()
	}
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (pi *GanjifaInteractor) NextTrick() string {
	pi.Game.NextTrick()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (pi *GanjifaInteractor) NextRound() string {
	pi.Game.ScoreRound()
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.advanceCpu()
	return pi.sp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *GanjifaInteractor) GetConfig() domain.GanjifaConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *GanjifaInteractor) Hint() string {
	return pi.sp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *GanjifaInteractor) ActionLog() string {
	return pi.sp.ActionLogOutput(pi.Game)
}

// advanceCpu 人間の手番になるまで CPU を自動進行させる。
//
// **入札フェーズは無い。**Ganjifa の切り札はディーラーの手札から自動で決まる
// ので、Préférence にあった入札ループはここには要らない。
func (pi *GanjifaInteractor) advanceCpu() {
	pi.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (pi *GanjifaInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.GanjifaPhase]{
		play:     domain.GanjifaPhasePlay,
		trickEnd: domain.GanjifaPhaseTrickEnd,
		roundEnd: domain.GanjifaPhaseRoundEnd,
		gameEnd:  domain.GanjifaPhaseGameEnd,
	})
}

// RestoreGanjifaInteractor deserialises JSON into a GanjifaInteractor.
func RestoreGanjifaInteractor(data []byte, sp presenter.GanjifaPresenter) (*GanjifaInteractor, error) {
	return restoreAndBuild[domain.Ganjifa](data, func(g *domain.Ganjifa) *GanjifaInteractor {
		return &GanjifaInteractor{GameBase: GameBase[interfaces.GanjifaGame]{Game: g}, sp: sp}
	})
}
