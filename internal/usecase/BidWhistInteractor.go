//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BidWhistInteractorIF Bid Whist インタラクターインタフェース
type BidWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BidWhistConfig) string
	// Bid ビッドする (tricks=ブックを超える目標数, direction=Uptown/Downtown/NoTrump)
	Bid(tricks, direction int) string
	// Pass パスする
	Pass() string
	// DeclareTrump 切り札スートを宣言する
	DeclareTrump(suit int) string
	// ExchangeKitty キティ交換 (6枚捨てる)
	ExchangeKitty(discardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BidWhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BidWhistInteractor Bid Whist インタラクタークラス
type BidWhistInteractor struct {
	GameBase[interfaces.BidWhistGame]
	fp presenter.BidWhistPresenter
}

// NewBidWhistInteractor コンストラクタ
func NewBidWhistInteractor(g interfaces.BidWhistGame, fp presenter.BidWhistPresenter) *BidWhistInteractor {
	mustNotNil("BidWhistInteractor", map[string]any{"g": g, "fp": fp})
	return &BidWhistInteractor{GameBase: GameBase[interfaces.BidWhistGame]{Game: g}, fp: fp}
}

// Reset ゲーム初期化
func (bi *BidWhistInteractor) Reset() string {
	bi.Game.Reset()
	bi.advanceCpu()
	return bi.fp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BidWhistInteractor) ResetWithConfig(cfg domain.BidWhistConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.fp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Bid ビッドする
func (bi *BidWhistInteractor) Bid(tricks, direction int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	if err := bi.Game.PlayerBid(tricks, direction); err != nil {
		return bi.fp.Output(bi.Game, err)
	}
	bi.advanceCpu()
	return bi.fp.Output(bi.Game, nil)
}

// Pass パスする
func (bi *BidWhistInteractor) Pass() string {
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	if err := bi.Game.PlayerPass(); err != nil {
		return bi.fp.Output(bi.Game, err)
	}
	bi.advanceCpu()
	return bi.fp.Output(bi.Game, nil)
}

// DeclareTrump 切り札スートを宣言する (落札者が人間のとき)
func (bi *BidWhistInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	if err := bi.Game.PlayerDeclareTrump(suit); err != nil {
		return bi.fp.Output(bi.Game, err)
	}
	// 人間の落札者が宣言したので、次は同じ人間がキティ交換を行う。自動進行は不要。
	return bi.fp.Output(bi.Game, nil)
}

// ExchangeKitty キティ交換 (6枚捨てる)
func (bi *BidWhistInteractor) ExchangeKitty(discardIndices []int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	if err := bi.Game.PlayerExchangeKitty(discardIndices); err != nil {
		return bi.fp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.fp.Output(bi.Game, nil)
}

// Play カードをプレイ
func (bi *BidWhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.fp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.fp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.fp.Output(bi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (bi *BidWhistInteractor) NextTrick() string {
	bi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	bi.Game.NextTrick()
	bi.runCpuTurns()
	return bi.fp.Output(bi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (bi *BidWhistInteractor) NextRound() string {
	bi.Game.ScoreRound()
	if out, blocked := guardGameEnd(bi.Game, bi.fp); blocked {
		return out
	}
	bi.Game.NextRound()
	bi.advanceCpu()
	return bi.fp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BidWhistInteractor) GetConfig() domain.BidWhistConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BidWhistInteractor) Hint() string {
	return bi.fp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BidWhistInteractor) ActionLog() string {
	return bi.fp.ActionLogOutput(bi.Game)
}

// advanceCpu ビッド → 切り札宣言 → キティ交換 → プレイ の順に CPU を自動進行させる
func (bi *BidWhistInteractor) advanceCpu() {
	runCpuBidsLoop(bi.Game, domain.BidWhistPhaseBid)
	bi.runCpuDeclareTrump()
	bi.runCpuExchange()
	bi.runCpuTurns()
}

// runCpuDeclareTrump 切り札宣言フェーズで CPU 落札者を自動実行する
func (bi *BidWhistInteractor) runCpuDeclareTrump() {
	if bi.Game.GetGameEndFlag() {
		return
	}
	if bi.Game.GetPhase() == domain.BidWhistPhaseTrumpDeclaration && !bi.Game.IsHumanDeclarerTurn() {
		bi.Game.CpuDeclareTrump()
	}
}

// runCpuExchange キティ交換フェーズで CPU 落札者を自動実行する
func (bi *BidWhistInteractor) runCpuExchange() {
	if bi.Game.GetGameEndFlag() {
		return
	}
	if bi.Game.GetPhase() == domain.BidWhistPhaseKittyExchange && !bi.Game.IsHumanDeclarerTurn() {
		bi.Game.CpuExchange()
	}
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する
func (bi *BidWhistInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.BidWhistPhase]{
		play:     domain.BidWhistPhasePlay,
		trickEnd: domain.BidWhistPhaseTrickEnd,
		roundEnd: domain.BidWhistPhaseRoundEnd,
		gameEnd:  domain.BidWhistPhaseGameEnd,
	})
}

// RestoreBidWhistInteractor deserialises JSON into a BidWhistInteractor.
func RestoreBidWhistInteractor(data []byte, fp presenter.BidWhistPresenter) (*BidWhistInteractor, error) {
	g, err := restoreGame[domain.BidWhist](data)
	if err != nil {
		return nil, err
	}
	return &BidWhistInteractor{GameBase: GameBase[interfaces.BidWhistGame]{Game: g}, fp: fp}, nil
}
