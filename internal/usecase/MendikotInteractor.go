//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MendikotInteractorIF メンディコットインタラクターインタフェース
type MendikotInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MendikotConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MendikotConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MendikotInteractor メンディコットインタラクタークラス
type MendikotInteractor struct {
	GameBase[interfaces.MendikotGame]
	mp presenter.MendikotPresenter
}

// NewMendikotInteractor コンストラクタ
func NewMendikotInteractor(m interfaces.MendikotGame, mp presenter.MendikotPresenter) *MendikotInteractor {
	mustNotNil("MendikotInteractor", map[string]any{"m": m, "mp": mp})
	return &MendikotInteractor{GameBase: GameBase[interfaces.MendikotGame]{Game: m}, mp: mp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **切り札を決める専用フェーズが無いので、進めるのはプレイだけ。** リードは
// 親の左隣なので、人間が親でなければ CPU が先に打つ。
func (mi *MendikotInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MendikotInteractor) ResetWithConfig(cfg domain.MendikotConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.mp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Play カードをプレイ
func (mi *MendikotInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.mp); blocked {
		return out
	}
	if err := mi.Game.PlayerPlay(cardIndex); err != nil {
		return mi.mp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// NextHand 次のハンドへ進む
func (mi *MendikotInteractor) NextHand() string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.NextHand()
	mi.runCpuTurns()
	return mi.mp.Output(mi.Game, nil)
}

// GiveUp 投了する
func (mi *MendikotInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(mi.Game, mi.mp); blocked {
		return out
	}
	mi.Game.GiveUp()
	return mi.mp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MendikotInteractor) GetConfig() domain.MendikotConfig { return mi.Game.GetConfig() }

// Hint ヒント取得
func (mi *MendikotInteractor) Hint() string { return mi.mp.HintOutput(mi.Game) }

// ActionLog 棋譜を出力する
func (mi *MendikotInteractor) ActionLog() string { return mi.mp.ActionLogOutput(mi.Game) }

// runCpuTurns 人間の手番になるかハンド／ゲームが終わるまで CPU を進める
func (mi *MendikotInteractor) runCpuTurns() {
	for turns := 0; !mi.Game.GetGameEndFlag() && !mi.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if mi.Game.GetPhase() != domain.MendikotPhasePlay {
			return
		}
		mi.Game.CpuPlay()
	}
}

// RestoreMendikotInteractor deserialises JSON into a MendikotInteractor.
func RestoreMendikotInteractor(data []byte, mp presenter.MendikotPresenter) (*MendikotInteractor, error) {
	return restoreAndBuild[domain.Mendikot](data, func(g *domain.Mendikot) *MendikotInteractor {
		return &MendikotInteractor{GameBase: GameBase[interfaces.MendikotGame]{Game: g}, mp: mp}
	})
}
