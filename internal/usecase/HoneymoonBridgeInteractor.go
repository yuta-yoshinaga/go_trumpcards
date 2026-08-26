//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HoneymoonBridgeInteractorIF ハネムーンブリッジインタラクターインタフェース
type HoneymoonBridgeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HoneymoonBridgeConfig) string
	// Bid 契約を宣言する
	Bid(level, suit int) string
	// Pass 降りる
	Pass() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のディールへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HoneymoonBridgeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HoneymoonBridgeInteractor ハネムーンブリッジインタラクタークラス
type HoneymoonBridgeInteractor struct {
	GameBase[interfaces.HoneymoonBridgeGame]
	sp presenter.HoneymoonBridgePresenter
}

// NewHoneymoonBridgeInteractor コンストラクタ
func NewHoneymoonBridgeInteractor(s interfaces.HoneymoonBridgeGame, sp presenter.HoneymoonBridgePresenter) *HoneymoonBridgeInteractor {
	mustNotNil("HoneymoonBridgeInteractor", map[string]any{"s": s, "sp": sp})
	return &HoneymoonBridgeInteractor{GameBase: GameBase[interfaces.HoneymoonBridgeGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (hi *HoneymoonBridgeInteractor) Reset() string {
	hi.Game.Reset()
	hi.advance()
	return hi.sp.Output(hi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HoneymoonBridgeInteractor) ResetWithConfig(cfg domain.HoneymoonBridgeConfig) string {
	return resetWithValidatedConfig(hi.Game, hi.sp, cfg, hi.Game.SetConfig, hi.Reset)
}

// Bid 契約を宣言する
func (hi *HoneymoonBridgeInteractor) Bid(level, suit int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.sp); blocked {
		return out
	}
	if err := hi.Game.PlayerBid(level, suit); err != nil {
		return hi.sp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.sp.Output(hi.Game, nil)
}

// Pass 降りる
func (hi *HoneymoonBridgeInteractor) Pass() string {
	if out, blocked := guardGameEnd(hi.Game, hi.sp); blocked {
		return out
	}
	if err := hi.Game.PlayerPass(); err != nil {
		return hi.sp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.sp.Output(hi.Game, nil)
}

// Play カードをプレイ
func (hi *HoneymoonBridgeInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(hi.Game, hi.sp); blocked {
		return out
	}
	if err := hi.Game.PlayerPlay(cardIndex); err != nil {
		return hi.sp.Output(hi.Game, err)
	}
	hi.advance()
	return hi.sp.Output(hi.Game, nil)
}

// NextRound 次のディールへ進む
func (hi *HoneymoonBridgeInteractor) NextRound() string {
	if out, blocked := guardGameEnd(hi.Game, hi.sp); blocked {
		return out
	}
	hi.Game.NextRound()
	hi.advance()
	return hi.sp.Output(hi.Game, nil)
}

// GiveUp 投了する
func (hi *HoneymoonBridgeInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(hi.Game, hi.sp); blocked {
		return out
	}
	hi.Game.GiveUp()
	return hi.sp.Output(hi.Game, nil)
}

// GetConfig 現在の設定を取得
func (hi *HoneymoonBridgeInteractor) GetConfig() domain.HoneymoonBridgeConfig {
	return hi.Game.GetConfig()
}

// Hint ヒント取得
func (hi *HoneymoonBridgeInteractor) Hint() string { return hi.sp.HintOutput(hi.Game) }

// ActionLog 棋譜を出力する
func (hi *HoneymoonBridgeInteractor) ActionLog() string { return hi.sp.ActionLogOutput(hi.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **引き合い・競り・本番の 3 段すべてを回す。** 引き合いと本番は同じ
// `IsHumanTurn` で判定できるが、競りだけは別の入口なので落とすと
// CPU の手番の盤面を返してしまう。ラウンド終了では止める（next は明示的に）。
func (hi *HoneymoonBridgeInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if hi.Game.GetGameEndFlag() {
			return
		}
		switch hi.Game.GetPhase() {
		case domain.HoneymoonBridgePhaseDraw, domain.HoneymoonBridgePhasePlay:
			if hi.Game.IsHumanTurn() {
				return
			}
			hi.Game.CpuPlay()
		case domain.HoneymoonBridgePhaseBid:
			if hi.Game.IsHumanBidTurn() {
				return
			}
			hi.Game.CpuBid()
		default:
			return
		}
	}
}

// RestoreHoneymoonBridgeInteractor deserialises JSON into a HoneymoonBridgeInteractor.
func RestoreHoneymoonBridgeInteractor(data []byte, sp presenter.HoneymoonBridgePresenter) (*HoneymoonBridgeInteractor, error) {
	return restoreAndBuild[domain.HoneymoonBridge](data, func(g *domain.HoneymoonBridge) *HoneymoonBridgeInteractor {
		return NewHoneymoonBridgeInteractor(g, sp)
	})
}
