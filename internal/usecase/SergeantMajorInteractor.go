//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SergeantMajorInteractorIF サージェントメジャーインタラクターインタフェース
type SergeantMajorInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SergeantMajorConfig) string
	// DeclareTrump 切り札を宣言する
	DeclareTrump(suit int) string
	// Discard キティのぶんを捨てる
	Discard(indices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SergeantMajorConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SergeantMajorInteractor サージェントメジャーインタラクタークラス
type SergeantMajorInteractor struct {
	GameBase[interfaces.SergeantMajorGame]
	sp presenter.SergeantMajorPresenter
}

// NewSergeantMajorInteractor コンストラクタ
func NewSergeantMajorInteractor(s interfaces.SergeantMajorGame, sp presenter.SergeantMajorPresenter) *SergeantMajorInteractor {
	mustNotNil("SergeantMajorInteractor", map[string]any{"s": s, "sp": sp})
	return &SergeantMajorInteractor{GameBase: GameBase[interfaces.SergeantMajorGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (si *SergeantMajorInteractor) Reset() string {
	si.Game.Reset()
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SergeantMajorInteractor) ResetWithConfig(cfg domain.SergeantMajorConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// DeclareTrump 切り札を宣言する
func (si *SergeantMajorInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDeclareTrump(suit); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// Discard キティのぶんを捨てる
func (si *SergeantMajorInteractor) Discard(indices []int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDiscard(indices); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// Play カードをプレイ
func (si *SergeantMajorInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンドへ進む
func (si *SergeantMajorInteractor) NextRound() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.advance()
	return si.sp.Output(si.Game, nil)
}

// GiveUp 投了する
func (si *SergeantMajorInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.GiveUp()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SergeantMajorInteractor) GetConfig() domain.SergeantMajorConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SergeantMajorInteractor) Hint() string { return si.sp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する
func (si *SergeantMajorInteractor) ActionLog() string { return si.sp.ActionLogOutput(si.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **宣言・捨て札・プレイの 3 段すべてを回す。** どれか 1 つで止めると、人間が
// 操作できない盤面を返してしまう。ラウンド終了では止める（next は明示的に）。
func (si *SergeantMajorInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		switch si.Game.GetPhase() {
		case domain.SergeantMajorPhaseTrump:
			if si.Game.IsHumanTrumpTurn() {
				return
			}
			si.Game.CpuDeclareTrump()
		case domain.SergeantMajorPhaseDiscard:
			if si.Game.IsHumanDiscardTurn() {
				return
			}
			si.Game.CpuDiscard()
		case domain.SergeantMajorPhasePlay:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
		default:
			return
		}
	}
}

// RestoreSergeantMajorInteractor deserialises JSON into a SergeantMajorInteractor.
func RestoreSergeantMajorInteractor(data []byte, sp presenter.SergeantMajorPresenter) (*SergeantMajorInteractor, error) {
	return restoreAndBuild[domain.SergeantMajor](data, func(g *domain.SergeantMajor) *SergeantMajorInteractor {
		return NewSergeantMajorInteractor(g, sp)
	})
}
