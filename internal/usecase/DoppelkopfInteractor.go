//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoppelkopfInteractorIF ドッペルコップのインタラクターインタフェース
type DoppelkopfInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.DoppelkopfConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Announce Re/Kontra を宣言する
	Announce() string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.DoppelkopfConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DoppelkopfInteractor ドッペルコップのインタラクタークラス
type DoppelkopfInteractor struct {
	GameBase[interfaces.DoppelkopfGame]
	dp presenter.DoppelkopfPresenter
}

// NewDoppelkopfInteractor コンストラクタ
func NewDoppelkopfInteractor(g interfaces.DoppelkopfGame, dp presenter.DoppelkopfPresenter) *DoppelkopfInteractor {
	mustNotNil("DoppelkopfInteractor", map[string]any{"g": g, "dp": dp})
	return &DoppelkopfInteractor{GameBase: GameBase[interfaces.DoppelkopfGame]{Game: g}, dp: dp}
}

// Reset ゲーム初期化
func (di *DoppelkopfInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (di *DoppelkopfInteractor) ResetWithConfig(cfg domain.DoppelkopfConfig) string {
	return resetWithValidatedConfig(di.Game, di.dp, cfg, di.Game.SetConfig, di.Reset)
}

// Play カードをプレイ
func (di *DoppelkopfInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerPlay(cardIndex)
	if err != nil {
		return di.dp.Output(di.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if di.Game.GetPhase() == domain.DoppelkopfPhaseTrickEnd {
		di.Game.ResolveTrick()
	}
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// Announce Re/Kontra を宣言する
func (di *DoppelkopfInteractor) Announce() string {
	if out, blocked := guardGameEnd(di.Game, di.dp); blocked {
		return out
	}
	if err := di.Game.PlayerAnnounce(); err != nil {
		return di.dp.Output(di.Game, err)
	}
	return di.dp.Output(di.Game, nil)
}

// NextTrick 次のトリックへ進む
func (di *DoppelkopfInteractor) NextTrick() string {
	di.Game.NextTrick()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (di *DoppelkopfInteractor) NextRound() string {
	di.Game.ScoreRound()
	if out, blocked := guardGameEnd(di.Game, di.dp); blocked {
		return out
	}
	di.Game.NextRound()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// GetConfig 現在の設定を取得
func (di *DoppelkopfInteractor) GetConfig() domain.DoppelkopfConfig {
	return di.Game.GetConfig()
}

// Hint ヒント取得
func (di *DoppelkopfInteractor) Hint() string {
	return di.dp.HintOutput(di.Game)
}

// ActionLog 棋譜を出力する
func (di *DoppelkopfInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (di *DoppelkopfInteractor) runCpuTurns() {
	runCpuTurnsLoop(di.Game, trickPhases[domain.DoppelkopfPhase]{
		play:     domain.DoppelkopfPhasePlay,
		trickEnd: domain.DoppelkopfPhaseTrickEnd,
		roundEnd: domain.DoppelkopfPhaseRoundEnd,
		gameEnd:  domain.DoppelkopfPhaseGameEnd,
	})
}

// RestoreDoppelkopfInteractor deserialises JSON into a DoppelkopfInteractor.
func RestoreDoppelkopfInteractor(data []byte, dp presenter.DoppelkopfPresenter) (*DoppelkopfInteractor, error) {
	return restoreAndBuild[domain.Doppelkopf](data, func(g *domain.Doppelkopf) *DoppelkopfInteractor {
		return &DoppelkopfInteractor{GameBase: GameBase[interfaces.DoppelkopfGame]{Game: g}, dp: dp}
	})
}
