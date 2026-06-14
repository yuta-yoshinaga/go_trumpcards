//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SuecaInteractorIF スエカのインタラクターインタフェース
type SuecaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SuecaConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SuecaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SuecaInteractor スエカのインタラクタークラス
type SuecaInteractor struct {
	GameBase[interfaces.SuecaGame]
	sp presenter.SuecaPresenter
}

// NewSuecaInteractor コンストラクタ
func NewSuecaInteractor(g interfaces.SuecaGame, sp presenter.SuecaPresenter) *SuecaInteractor {
	mustNotNil("SuecaInteractor", map[string]any{"g": g, "sp": sp})
	return &SuecaInteractor{GameBase: GameBase[interfaces.SuecaGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SuecaInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SuecaInteractor) ResetWithConfig(cfg domain.SuecaConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Play カードをプレイ
func (si *SuecaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(cardIndex)
	if err != nil {
		return si.sp.Output(si.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if si.Game.GetPhase() == domain.SuecaPhaseTrickEnd {
		si.Game.ResolveTrick()
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextTrick 次のトリックへ進む
func (si *SuecaInteractor) NextTrick() string {
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SuecaInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SuecaInteractor) GetConfig() domain.SuecaConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SuecaInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SuecaInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (si *SuecaInteractor) runCpuTurns() {
	runCpuTurnsLoop(si.Game, trickPhases[domain.SuecaPhase]{
		play:     domain.SuecaPhasePlay,
		trickEnd: domain.SuecaPhaseTrickEnd,
		roundEnd: domain.SuecaPhaseRoundEnd,
		gameEnd:  domain.SuecaPhaseGameEnd,
	})
}

// RestoreSuecaInteractor deserialises JSON into a SuecaInteractor.
func RestoreSuecaInteractor(data []byte, sp presenter.SuecaPresenter) (*SuecaInteractor, error) {
	return restoreAndBuild[domain.Sueca](data, func(g *domain.Sueca) *SuecaInteractor {
		return &SuecaInteractor{GameBase: GameBase[interfaces.SuecaGame]{Game: g}, sp: sp}
	})
}
