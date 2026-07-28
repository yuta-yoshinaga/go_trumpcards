//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RookInteractorIF ルーク(Rook)インタラクターインタフェース
type RookInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.RookConfig) string
	// Bid ビッドする
	Bid(bid int) string
	// Pass パスする
	Pass() string
	// ExchangeNest ネスト交換 (5枚捨てて切り札色を宣言)
	ExchangeNest(discardIndices []int, trumpColor int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.RookConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RookInteractor ルーク(Rook)インタラクタークラス
type RookInteractor struct {
	GameBase[interfaces.RookGame]
	fp presenter.RookPresenter
}

// NewRookInteractor コンストラクタ
func NewRookInteractor(g interfaces.RookGame, fp presenter.RookPresenter) *RookInteractor {
	mustNotNil("RookInteractor", map[string]any{"g": g, "fp": fp})
	return &RookInteractor{GameBase: GameBase[interfaces.RookGame]{Game: g}, fp: fp}
}

// Reset ゲーム初期化
func (fi *RookInteractor) Reset() string {
	fi.Game.Reset()
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (fi *RookInteractor) ResetWithConfig(cfg domain.RookConfig) string {
	return resetWithValidatedConfig(fi.Game, fi.fp, cfg, fi.Game.SetConfig, fi.Reset)
}

// Bid ビッドする
func (fi *RookInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerBid(bid); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// Pass パスする
func (fi *RookInteractor) Pass() string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerPass(); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// ExchangeNest ネスト交換 (5枚捨てて切り札色を宣言)
func (fi *RookInteractor) ExchangeNest(discardIndices []int, trumpColor int) string {
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerExchangeNest(discardIndices, trumpColor); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// Play カードをプレイ
func (fi *RookInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(fi.Game, fi.fp); blocked {
		return out
	}
	if err := fi.Game.PlayerPlay(cardIndex); err != nil {
		return fi.fp.Output(fi.Game, err)
	}
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (fi *RookInteractor) NextTrick() string {
	fi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	fi.Game.NextTrick()
	fi.runCpuTurns()
	return fi.fp.Output(fi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (fi *RookInteractor) NextRound() string {
	fi.Game.ScoreRound()
	if out, blocked := guardGameEnd(fi.Game, fi.fp); blocked {
		return out
	}
	fi.Game.NextRound()
	fi.advanceCpu()
	return fi.fp.Output(fi.Game, nil)
}

// GetConfig 現在の設定を取得
func (fi *RookInteractor) GetConfig() domain.RookConfig {
	return fi.Game.GetConfig()
}

// Hint ヒント取得
func (fi *RookInteractor) Hint() string {
	return fi.fp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *RookInteractor) ActionLog() string {
	return fi.fp.ActionLogOutput(fi.Game)
}

// advanceCpu ビッド → ネスト交換 → プレイ の順に CPU を自動進行させる
func (fi *RookInteractor) advanceCpu() {
	runCpuBidsLoop(fi.Game, domain.RookPhaseBid)
	fi.runCpuExchange()
	fi.runCpuTurns()
}

// runCpuExchange ネスト交換フェーズで CPU 落札者を自動実行する
func (fi *RookInteractor) runCpuExchange() {
	if fi.Game.GetGameEndFlag() {
		return
	}
	if fi.Game.GetPhase() == domain.RookPhaseNestExchange {
		fi.Game.CpuExchange()
	}
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する
func (fi *RookInteractor) runCpuTurns() {
	runCpuTurnsLoop(fi.Game, trickPhases[domain.RookPhase]{
		play:     domain.RookPhasePlay,
		trickEnd: domain.RookPhaseTrickEnd,
		roundEnd: domain.RookPhaseRoundEnd,
		gameEnd:  domain.RookPhaseGameEnd,
	})
}

// RestoreRookInteractor deserialises JSON into a RookInteractor.
func RestoreRookInteractor(data []byte, fp presenter.RookPresenter) (*RookInteractor, error) {
	g, err := restoreGame[domain.Rook](data)
	if err != nil {
		return nil, err
	}
	return &RookInteractor{GameBase: GameBase[interfaces.RookGame]{Game: g}, fp: fp}, nil
}
