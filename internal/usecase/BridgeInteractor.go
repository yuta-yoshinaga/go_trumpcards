//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BridgeInteractorIF ブリッジインタラクターインタフェース
type BridgeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BridgeConfig) string
	// Bid ビッドする (bidType: 0=Pass, 1=Normal, 2=Double, 3=Redouble)
	Bid(bidType int, level int, suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BridgeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BridgeInteractor ブリッジインタラクタークラス
type BridgeInteractor struct {
	GameBase[interfaces.BridgeGame]
	bp presenter.BridgePresenter
}

// NewBridgeInteractor コンストラクタ
func NewBridgeInteractor(b interfaces.BridgeGame, bp presenter.BridgePresenter) *BridgeInteractor {
	mustNotNil("BridgeInteractor", map[string]any{"b": b, "bp": bp})
	return &BridgeInteractor{GameBase: GameBase[interfaces.BridgeGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BridgeInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BridgeInteractor) ResetWithConfig(cfg domain.BridgeConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bid ビッドする
func (bi *BridgeInteractor) Bid(bidType int, level int, suit int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerBid(bidType, level, suit)
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Play カードをプレイ
func (bi *BridgeInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (bi *BridgeInteractor) NextTrick() string {
	bi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextTrick()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (bi *BridgeInteractor) NextRound() string {
	bi.Game.ScoreRound()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextRound()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BridgeInteractor) GetConfig() domain.BridgeConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BridgeInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BridgeInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuBids ビッドフェーズでCPUを自動実行する
func (bi *BridgeInteractor) runCpuBids() {
	runCpuBidsLoop(bi.Game, domain.BridgePhaseBid)
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (bi *BridgeInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.BridgePhase]{
		play:     domain.BridgePhasePlay,
		trickEnd: domain.BridgePhaseTrickEnd,
		roundEnd: domain.BridgePhaseRoundEnd,
		gameEnd:  domain.BridgePhaseGameEnd,
	})
}

// RestoreBridgeInteractor deserialises JSON into a BridgeInteractor.
func RestoreBridgeInteractor(data []byte, bp presenter.BridgePresenter) (*BridgeInteractor, error) {
	return restoreAndBuild[domain.Bridge](data, func(g *domain.Bridge) *BridgeInteractor {
		return &BridgeInteractor{GameBase: GameBase[interfaces.BridgeGame]{Game: g}, bp: bp}
	})
}
