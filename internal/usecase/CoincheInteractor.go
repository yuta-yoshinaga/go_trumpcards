//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CoincheInteractorIF コワンシュインタラクターインタフェース
type CoincheInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CoincheConfig) string
	// Bid 目標点と切り札スートを宣言する
	Bid(points, suit int) string
	// Pass 競りでパスする
	Pass() string
	// Coinche 守備側が倍化する
	Coinche() string
	// Surcoinche 宣言側が再倍化する
	Surcoinche() string
	// DeclineDouble 倍化せずに進める
	DeclineDouble() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CoincheConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CoincheInteractor コワンシュインタラクタークラス
type CoincheInteractor struct {
	GameBase[interfaces.CoincheGame]
	bp presenter.CoinchePresenter
}

// NewCoincheInteractor コンストラクタ
func NewCoincheInteractor(b interfaces.CoincheGame, bp presenter.CoinchePresenter) *CoincheInteractor {
	mustNotNil("CoincheInteractor", map[string]any{"b": b, "bp": bp})
	return &CoincheInteractor{GameBase: GameBase[interfaces.CoincheGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *CoincheInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *CoincheInteractor) ResetWithConfig(cfg domain.CoincheConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bid 目標点と切り札スートを宣言する
func (bi *CoincheInteractor) Bid(points, suit int) string {
	return bi.afterBidAction(func() error { return bi.Game.PlayerBid(points, suit) })
}

// Pass 競りでパスする
func (bi *CoincheInteractor) Pass() string {
	return bi.afterBidAction(bi.Game.PlayerPassBid)
}

// Coinche 守備側が倍化する
func (bi *CoincheInteractor) Coinche() string {
	return bi.afterBidAction(bi.Game.PlayerCoinche)
}

// Surcoinche 宣言側が再倍化する
func (bi *CoincheInteractor) Surcoinche() string {
	return bi.afterBidAction(bi.Game.PlayerSurcoinche)
}

// DeclineDouble 倍化せずに進める
func (bi *CoincheInteractor) DeclineDouble() string {
	return bi.afterBidAction(bi.Game.PlayerDeclineDouble)
}

// afterBidAction は競り/倍化の 1 操作を適用し、続く CPU を回して出力する。
//
// **競りと倍化は同じ後始末を要る。** 別々に書くと、片方だけ CPU を回し忘れて
// 「自分の手番でないのに盤が止まる」形になる。
func (bi *CoincheInteractor) afterBidAction(apply func() error) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	if err := apply(); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Play カードをプレイ
func (bi *CoincheInteractor) Play(cardIndex int) string {
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
func (bi *CoincheInteractor) NextTrick() string {
	bi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextTrick()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (bi *CoincheInteractor) NextRound() string {
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
func (bi *CoincheInteractor) GetConfig() domain.CoincheConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *CoincheInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *CoincheInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuBids ビッドフェーズでCPUを自動実行する (PickUp と CallTrump)
func (bi *CoincheInteractor) runCpuBids() {
	for i := 0; i < MaxCpuIterations; i++ {
		if bi.Game.GetGameEndFlag() {
			return
		}
		phase := bi.Game.GetPhase()
		switch phase {
		case domain.CoinchePhaseBid:
			if bi.Game.IsHumanBidTurn() {
				return
			}
			bi.Game.CpuBid()
		case domain.CoinchePhaseDouble:
			if bi.Game.IsHumanTurn() {
				return
			}
			bi.Game.CpuDouble()
		default:
			return
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (bi *CoincheInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.CoinchePhase]{
		play:     domain.CoinchePhasePlay,
		trickEnd: domain.CoinchePhaseTrickEnd,
		roundEnd: domain.CoinchePhaseRoundEnd,
		gameEnd:  domain.CoinchePhaseGameEnd,
	})
}

// RestoreCoincheInteractor deserialises JSON into a CoincheInteractor.
func RestoreCoincheInteractor(data []byte, bp presenter.CoinchePresenter) (*CoincheInteractor, error) {
	b, err := restoreGame[domain.Coinche](data)
	if err != nil {
		return nil, err
	}
	return &CoincheInteractor{GameBase: GameBase[interfaces.CoincheGame]{Game: b}, bp: bp}, nil
}
