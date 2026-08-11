//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MinibridgeInteractorIF ミニブリッジインタラクターインタフェース
type MinibridgeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MinibridgeConfig) string
	// Contract 契約を選ぶ
	Contract(level, suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のディールへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MinibridgeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MinibridgeInteractor ミニブリッジインタラクタークラス
type MinibridgeInteractor struct {
	GameBase[interfaces.MinibridgeGame]
	sp presenter.MinibridgePresenter
}

// NewMinibridgeInteractor コンストラクタ
func NewMinibridgeInteractor(s interfaces.MinibridgeGame, sp presenter.MinibridgePresenter) *MinibridgeInteractor {
	mustNotNil("MinibridgeInteractor", map[string]any{"s": s, "sp": sp})
	return &MinibridgeInteractor{GameBase: GameBase[interfaces.MinibridgeGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (mi *MinibridgeInteractor) Reset() string {
	mi.Game.Reset()
	mi.advance()
	return mi.sp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MinibridgeInteractor) ResetWithConfig(cfg domain.MinibridgeConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.sp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Contract 契約を選ぶ
func (mi *MinibridgeInteractor) Contract(level, suit int) string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	if err := mi.Game.PlayerSelectContract(level, suit); err != nil {
		return mi.sp.Output(mi.Game, err)
	}
	mi.advance()
	return mi.sp.Output(mi.Game, nil)
}

// Play カードをプレイ
func (mi *MinibridgeInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(mi.Game, mi.sp); blocked {
		return out
	}
	if err := mi.Game.PlayerPlay(cardIndex); err != nil {
		return mi.sp.Output(mi.Game, err)
	}
	mi.advance()
	return mi.sp.Output(mi.Game, nil)
}

// NextRound 次のディールへ進む
func (mi *MinibridgeInteractor) NextRound() string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	mi.Game.NextRound()
	mi.advance()
	return mi.sp.Output(mi.Game, nil)
}

// GiveUp 投了する
func (mi *MinibridgeInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	mi.Game.GiveUp()
	return mi.sp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MinibridgeInteractor) GetConfig() domain.MinibridgeConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *MinibridgeInteractor) Hint() string { return mi.sp.HintOutput(mi.Game) }

// ActionLog 棋譜を出力する
func (mi *MinibridgeInteractor) ActionLog() string { return mi.sp.ActionLogOutput(mi.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **ダミーの手番は人間デクレアラーの出番。** `IsHumanTurn` がそれを含めて
// 真を返すので、ここで CPU に打たせてしまうとダミーを人間が操作できなくなる。
// ディール終了では止める（next は明示的に）。
func (mi *MinibridgeInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if mi.Game.GetGameEndFlag() {
			return
		}
		switch mi.Game.GetPhase() {
		case domain.MinibridgePhaseContract:
			if mi.Game.IsHumanContractTurn() {
				return
			}
			mi.Game.CpuSelectContract()
		case domain.MinibridgePhasePlay:
			if mi.Game.IsHumanTurn() {
				return
			}
			mi.Game.CpuPlay()
		default:
			return
		}
	}
}

// RestoreMinibridgeInteractor deserialises JSON into a MinibridgeInteractor.
func RestoreMinibridgeInteractor(data []byte, sp presenter.MinibridgePresenter) (*MinibridgeInteractor, error) {
	return restoreAndBuild[domain.Minibridge](data, func(g *domain.Minibridge) *MinibridgeInteractor {
		return NewMinibridgeInteractor(g, sp)
	})
}
