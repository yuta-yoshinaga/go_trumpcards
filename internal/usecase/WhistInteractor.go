package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WhistInteractorIF ホイストインタラクターインタフェース
type WhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.WhistConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.WhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// WhistInteractor ホイストインタラクタークラス
type WhistInteractor struct {
	GameBase[interfaces.WhistGame]
	wp presenter.WhistPresenter
}

// NewWhistInteractor コンストラクタ
func NewWhistInteractor(w interfaces.WhistGame, wp presenter.WhistPresenter) *WhistInteractor {
	mustNotNil("WhistInteractor", map[string]any{"w": w, "wp": wp})
	return &WhistInteractor{GameBase: GameBase[interfaces.WhistGame]{Game: w}, wp: wp}
}

// Reset ゲーム初期化
func (wi *WhistInteractor) Reset() string {
	wi.Game.Reset()
	wi.runCpuTurns()
	return wi.wp.Output(wi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (wi *WhistInteractor) ResetWithConfig(cfg domain.WhistConfig) string {
	return resetWithValidatedConfig(wi.Game, wi.wp, cfg, wi.Game.SetConfig, wi.Reset)
}

// Play カードをプレイ
func (wi *WhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(wi.Game, wi.wp); blocked {
		return out
	}
	err := wi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return wi.wp.Output(wi.Game, err)
	}
	wi.runCpuTurns()
	return wi.wp.Output(wi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (wi *WhistInteractor) NextTrick() string {
	wi.Game.NextTrick()
	wi.runCpuTurns()
	return wi.wp.Output(wi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (wi *WhistInteractor) NextRound() string {
	wi.Game.ScoreRound()
	if out, blocked := guardGameEnd(wi.Game, wi.wp); blocked {
		return out
	}
	wi.Game.NextRound()
	wi.runCpuTurns()
	return wi.wp.Output(wi.Game, nil)
}

// GetConfig 現在の設定を取得
func (wi *WhistInteractor) GetConfig() domain.WhistConfig {
	return wi.Game.GetConfig()
}

// Hint ヒント取得
func (wi *WhistInteractor) Hint() string {
	return wi.wp.HintOutput(wi.Game)
}

// ActionLog 棋譜を出力する
func (wi *WhistInteractor) ActionLog() string {
	return wi.wp.ActionLogOutput(wi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (wi *WhistInteractor) runCpuTurns() {
	runCpuTurnsLoop(wi.Game, trickPhases[domain.WhistPhase]{
		play:     domain.WhistPhasePlay,
		trickEnd: domain.WhistPhaseTrickEnd,
		roundEnd: domain.WhistPhaseRoundEnd,
		gameEnd:  domain.WhistPhaseGameEnd,
	})
}

// RestoreWhistInteractor deserialises JSON into a WhistInteractor.
func RestoreWhistInteractor(data []byte, wp presenter.WhistPresenter) (*WhistInteractor, error) {
	return restoreAndBuild[domain.Whist](data, func(g *domain.Whist) *WhistInteractor {
		return &WhistInteractor{GameBase: GameBase[interfaces.WhistGame]{Game: g}, wp: wp}
	})
}
