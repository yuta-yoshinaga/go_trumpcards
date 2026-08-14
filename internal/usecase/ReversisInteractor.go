//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ReversisInteractorIF レヴェルシインタラクターインタフェース
type ReversisInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ReversisConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ReversisConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ReversisInteractor レヴェルシインタラクタークラス
type ReversisInteractor struct {
	GameBase[interfaces.ReversisGame]
	rp presenter.ReversisPresenter
}

// NewReversisInteractor コンストラクタ
func NewReversisInteractor(r interfaces.ReversisGame, rp presenter.ReversisPresenter) *ReversisInteractor {
	mustNotNil("ReversisInteractor", map[string]any{"r": r, "rp": rp})
	return &ReversisInteractor{GameBase: GameBase[interfaces.ReversisGame]{Game: r}, rp: rp}
}

// Reset ゲーム初期化
func (ri *ReversisInteractor) Reset() string {
	ri.Game.Reset()
	ri.runCpuTurns()
	return ri.rp.Output(ri.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ri *ReversisInteractor) ResetWithConfig(cfg domain.ReversisConfig) string {
	return resetWithValidatedConfig(ri.Game, ri.rp, cfg, ri.Game.SetConfig, ri.Reset)
}

// Play カードをプレイ
func (ri *ReversisInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ri.Game, ri.rp); blocked {
		return out
	}
	if err := ri.Game.PlayerPlay(cardIndex); err != nil {
		return ri.rp.Output(ri.Game, err)
	}
	ri.runCpuTurns()
	return ri.rp.Output(ri.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ri *ReversisInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	ri.Game.NextRound()
	ri.runCpuTurns()
	return ri.rp.Output(ri.Game, nil)
}

// GiveUp 投了する
func (ri *ReversisInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	ri.Game.GiveUp()
	return ri.rp.Output(ri.Game, nil)
}

// GetConfig 現在の設定を取得
func (ri *ReversisInteractor) GetConfig() domain.ReversisConfig { return ri.Game.GetConfig() }

// Hint ヒント取得
func (ri *ReversisInteractor) Hint() string { return ri.rp.HintOutput(ri.Game) }

// ActionLog 棋譜を出力する
func (ri *ReversisInteractor) ActionLog() string { return ri.rp.ActionLogOutput(ri.Game) }

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める。
//
// **ラウンド終了では必ず止まる。** 誰がプールを取ったかを見せずに次を配って
// しまうと、このゲームの唯一の見どころが飛ぶ。
func (ri *ReversisInteractor) runCpuTurns() {
	for turns := 0; !ri.Game.GetGameEndFlag() && !ri.Game.IsHumanTurn(); turns++ {
		// 進まない CpuPlay でハングしないための上限 (#4607 と同じ理由)。
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ri.Game.GetPhase() != domain.ReversisPhasePlay {
			return
		}
		ri.Game.CpuPlay()
	}
}

// RestoreReversisInteractor deserialises JSON into a ReversisInteractor.
func RestoreReversisInteractor(data []byte, rp presenter.ReversisPresenter) (*ReversisInteractor, error) {
	return restoreAndBuild[domain.Reversis](data, func(g *domain.Reversis) *ReversisInteractor {
		return &ReversisInteractor{GameBase: GameBase[interfaces.ReversisGame]{Game: g}, rp: rp}
	})
}
