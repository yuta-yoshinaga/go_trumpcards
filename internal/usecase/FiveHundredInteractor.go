//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FiveHundredInteractorIF 500インタラクターインタフェース
type FiveHundredInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FiveHundredConfig) string
	// Bid ビッドする
	Bid(kind domain.FiveHundredContractKind, tricks, suit int) string
	// Pass パスする
	Pass() string
	// ExchangeKitty キティ交換 (3枚捨てる)
	ExchangeKitty(discardIndices []int) string
	// Play カードをプレイ (jokerSuit はNTでジョーカーをリードする際の指名スート)
	Play(cardIndex, jokerSuit int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FiveHundredConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FiveHundredInteractor 500インタラクタークラス
type FiveHundredInteractor struct {
	GameBase[interfaces.FiveHundredGame]
	fp presenter.FiveHundredPresenter
}

// NewFiveHundredInteractor コンストラクタ
func NewFiveHundredInteractor(g interfaces.FiveHundredGame, fp presenter.FiveHundredPresenter) *FiveHundredInteractor {
	mustNotNil("FiveHundredInteractor", map[string]any{"g": g, "fp": fp})
	return &FiveHundredInteractor{GameBase: GameBase[interfaces.FiveHundredGame]{Game: g}, fp: fp}
}

// Reset ゲーム初期化
func (fi *FiveHundredInteractor) Reset() string {
	fi.Game.Reset()
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (fi *FiveHundredInteractor) ResetWithConfig(cfg domain.FiveHundredConfig) string {
	return resetWithValidatedConfig(fi.Game, fi.fp, cfg, fi.Game.SetConfig, fi.Reset)
}

// Bid ビッドする
func (fi *FiveHundredInteractor) Bid(kind domain.FiveHundredContractKind, tricks, suit int) string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerBid(kind, tricks, suit); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// Pass パスする
func (fi *FiveHundredInteractor) Pass() string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerPass(); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// ExchangeKitty キティ交換 (3枚捨てる)
func (fi *FiveHundredInteractor) ExchangeKitty(discardIndices []int) string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerExchangeKitty(discardIndices); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// Play カードをプレイ
func (fi *FiveHundredInteractor) Play(cardIndex, jokerSuit int) string {
	if out, blocked := guardNotPlayable(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerPlay(cardIndex, jokerSuit); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (fi *FiveHundredInteractor) NextTrick() string {
	fi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	fi.Game.NextTrick()
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (fi *FiveHundredInteractor) NextRound() string {
	fi.Game.ScoreRound()
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	fi.Game.NextRound()
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// GetConfig 現在の設定を取得
func (fi *FiveHundredInteractor) GetConfig() domain.FiveHundredConfig {
	return fi.Game.GetConfig()
}

// Hint ヒント取得
func (fi *FiveHundredInteractor) Hint() string {
	return fi.fp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *FiveHundredInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// advanceCpu ビッド → キティ交換 → プレイ の順に CPU を自動進行させる
func (fi *FiveHundredInteractor) advanceCpu() {
	runCpuBidsLoop(fi.Game, domain.FiveHundredPhaseBid)
	fi.runCpuExchange()
	fi.runCpuTurns()
}

// runCpuExchange キティ交換フェーズで CPU 落札者を自動実行する
func (fi *FiveHundredInteractor) runCpuExchange() {
	if fi.Game.GetGameEndFlag() {
		return
	}
	if fi.Game.GetPhase() == domain.FiveHundredPhaseKittyExchange {
		fi.Game.CpuExchange()
	}
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する
func (fi *FiveHundredInteractor) runCpuTurns() {
	runCpuTurnsLoop(fi.Game, trickPhases[domain.FiveHundredPhase]{
		play:     domain.FiveHundredPhasePlay,
		trickEnd: domain.FiveHundredPhaseTrickEnd,
		roundEnd: domain.FiveHundredPhaseRoundEnd,
		gameEnd:  domain.FiveHundredPhaseGameEnd,
	})
}

// RestoreFiveHundredInteractor deserialises JSON into a FiveHundredInteractor.
func RestoreFiveHundredInteractor(data []byte, fp presenter.FiveHundredPresenter) (*FiveHundredInteractor, error) {
	g, err := restoreGame[domain.FiveHundred](data)
	if err != nil {
		return nil, err
	}
	return &FiveHundredInteractor{GameBase: GameBase[interfaces.FiveHundredGame]{Game: g}, fp: fp}, nil
}
