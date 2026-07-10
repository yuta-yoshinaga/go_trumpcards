package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CatchTenInteractorIF Catch the Ten インタラクターインタフェース
type CatchTenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CatchTenConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CatchTenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CatchTenInteractor Catch the Ten インタラクタークラス
type CatchTenInteractor struct {
	GameBase[interfaces.CatchTenGame]
	cp presenter.CatchTenPresenter
}

// NewCatchTenInteractor コンストラクタ
func NewCatchTenInteractor(g interfaces.CatchTenGame, cp presenter.CatchTenPresenter) *CatchTenInteractor {
	mustNotNil("CatchTenInteractor", map[string]any{"g": g, "cp": cp})
	return &CatchTenInteractor{GameBase: GameBase[interfaces.CatchTenGame]{Game: g}, cp: cp}
}

// Reset ゲーム初期化
func (ci *CatchTenInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CatchTenInteractor) ResetWithConfig(cfg domain.CatchTenConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *CatchTenInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.cp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	// If the human played the last card of the trick, resolve it here; when a
	// CPU plays the last card the trick is resolved inside runCpuTurns' loop.
	if ci.Game.GetPhase() == domain.CatchTenPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *CatchTenInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む (ラウンドのスコアリングはトリック解決時に自動実行済み)
func (ci *CatchTenInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CatchTenInteractor) GetConfig() domain.CatchTenConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *CatchTenInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CatchTenInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (ci *CatchTenInteractor) runCpuTurns() {
	runCpuTurnsLoop(ci.Game, trickPhases[domain.CatchTenPhase]{
		play:     domain.CatchTenPhasePlay,
		trickEnd: domain.CatchTenPhaseTrickEnd,
		roundEnd: domain.CatchTenPhaseRoundEnd,
		gameEnd:  domain.CatchTenPhaseGameEnd,
	})
}

// RestoreCatchTenInteractor deserialises JSON into a CatchTenInteractor.
func RestoreCatchTenInteractor(data []byte, cp presenter.CatchTenPresenter) (*CatchTenInteractor, error) {
	return restoreAndBuild[domain.CatchTen](data, func(g *domain.CatchTen) *CatchTenInteractor {
		return &CatchTenInteractor{GameBase: GameBase[interfaces.CatchTenGame]{Game: g}, cp: cp}
	})
}
