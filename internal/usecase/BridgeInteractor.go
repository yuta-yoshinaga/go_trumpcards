package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BridgeInteractorIF ブリッジインタラクターインタフェース
type BridgeInteractorIF interface {
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
	b  interfaces.BridgeGame
	bp presenter.BridgePresenter
}

// NewBridgeInteractor コンストラクタ
func NewBridgeInteractor(b interfaces.BridgeGame, bp presenter.BridgePresenter) *BridgeInteractor {
	mustNotNil("BridgeInteractor", map[string]any{"b": b, "bp": bp})
	return &BridgeInteractor{b: b, bp: bp}
}

// Reset ゲーム初期化
func (bi *BridgeInteractor) Reset() string {
	bi.b.Reset()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.b, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BridgeInteractor) ResetWithConfig(cfg domain.BridgeConfig) string {
	return resetWithValidatedConfig(bi.b, bi.bp, cfg, bi.b.SetConfig, bi.Reset)
}

// Bid ビッドする
func (bi *BridgeInteractor) Bid(bidType int, level int, suit int) string {
	if out, blocked := guardGameEnd(bi.b, bi.bp); blocked {
		return out
	}
	err := bi.b.PlayerBid(bidType, level, suit)
	if err != nil {
		return bi.bp.Output(bi.b, err)
	}
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.b, nil)
}

// Play カードをプレイ
func (bi *BridgeInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.b, bi.bp); blocked {
		return out
	}
	err := bi.b.PlayerPlay(cardIndex)
	if err != nil {
		return bi.bp.Output(bi.b, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.b, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (bi *BridgeInteractor) NextTrick() string {
	bi.b.ResolveTrick()
	if out, blocked := guardGameEnd(bi.b, bi.bp); blocked {
		return out
	}
	bi.b.NextTrick()
	bi.runCpuTurns()
	return bi.bp.Output(bi.b, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (bi *BridgeInteractor) NextRound() string {
	bi.b.ScoreRound()
	if out, blocked := guardGameEnd(bi.b, bi.bp); blocked {
		return out
	}
	bi.b.NextRound()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.b, nil)
}

// GetConfig 現在の設定を取得
func (bi *BridgeInteractor) GetConfig() domain.BridgeConfig {
	return bi.b.GetConfig()
}

// Hint ヒント取得
func (bi *BridgeInteractor) Hint() string {
	return bi.bp.HintOutput(bi.b)
}

// ActionLog 棋譜を出力する
func (bi *BridgeInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.b)
}

// runCpuBids ビッドフェーズでCPUを自動実行する
func (bi *BridgeInteractor) runCpuBids() {
	runCpuBidsLoop(bi.b, domain.BridgePhaseBid)
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (bi *BridgeInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.b, trickPhases[domain.BridgePhase]{
		play:     domain.BridgePhasePlay,
		trickEnd: domain.BridgePhaseTrickEnd,
		roundEnd: domain.BridgePhaseRoundEnd,
		gameEnd:  domain.BridgePhaseGameEnd,
	})
}

// Snapshot serialises the game state to JSON for KV persistence.
func (bi *BridgeInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(bi.b)
}

// RestoreBridgeInteractor deserialises JSON into a BridgeInteractor.
func RestoreBridgeInteractor(data []byte, bp presenter.BridgePresenter) (*BridgeInteractor, error) {
	b, err := restoreGame[domain.Bridge](data)
	if err != nil {
		return nil, err
	}
	return &BridgeInteractor{b: b, bp: bp}, nil
}
