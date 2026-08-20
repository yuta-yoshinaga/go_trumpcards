//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MusInteractorIF ムスのインタラクターインタフェース
type MusInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.MusConfig) string
	// Mus Mus(true)/Corte(false)を宣言する
	Mus(mus bool) string
	// Discard 交換する札を選ぶ
	Discard(indices []int) string
	// Bet 賭けアクションを実行する
	Bet(action, amount int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.MusConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MusInteractor ムスのインタラクタークラス
type MusInteractor struct {
	GameBase[interfaces.MusGame]
	sp presenter.MusPresenter
}

// NewMusInteractor コンストラクタ
func NewMusInteractor(g interfaces.MusGame, sp presenter.MusPresenter) *MusInteractor {
	mustNotNil("MusInteractor", map[string]any{"g": g, "sp": sp})
	return &MusInteractor{GameBase: GameBase[interfaces.MusGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (mi *MusInteractor) Reset() string {
	mi.Game.Reset()
	mi.runCpuTurns()
	return mi.sp.Output(mi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (mi *MusInteractor) ResetWithConfig(cfg domain.MusConfig) string {
	return resetWithValidatedConfig(mi.Game, mi.sp, cfg, mi.Game.SetConfig, mi.Reset)
}

// Mus Mus(true)/Corte(false)を宣言する
func (mi *MusInteractor) Mus(mus bool) string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	if err := mi.Game.PlayerMus(mus); err != nil {
		return mi.sp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.sp.Output(mi.Game, nil)
}

// Discard 交換する札を選ぶ
func (mi *MusInteractor) Discard(indices []int) string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	if err := mi.Game.PlayerDiscard(indices); err != nil {
		return mi.sp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.sp.Output(mi.Game, nil)
}

// Bet 賭けアクションを実行する
func (mi *MusInteractor) Bet(action, amount int) string {
	if out, blocked := guardGameEnd(mi.Game, mi.sp); blocked {
		return out
	}
	if err := mi.Game.PlayerBet(action, amount); err != nil {
		return mi.sp.Output(mi.Game, err)
	}
	mi.runCpuTurns()
	return mi.sp.Output(mi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (mi *MusInteractor) NextRound() string {
	mi.Game.NextRound()
	mi.runCpuTurns()
	return mi.sp.Output(mi.Game, nil)
}

// GetConfig 現在の設定を取得
func (mi *MusInteractor) GetConfig() domain.MusConfig {
	return mi.Game.GetConfig()
}

// Hint ヒント取得
func (mi *MusInteractor) Hint() string {
	return mi.sp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MusInteractor) ActionLog() string {
	return mi.sp.ActionLogOutput(mi.Game)
}

// runCpuTurns ゲーム終了・人間の手番・ラウンド終了になるまでCPUターンを実行する。
// Mus/Discard/Grande/Chica/Pares/Juego フェーズの CPU アクションを自動進行し、
// Showdown を解決し、RoundEnd で待機する。
func (mi *MusInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if mi.Game.GetGameEndFlag() {
			return
		}
		switch mi.Game.GetPhase() {
		case domain.MusPhaseMus, domain.MusPhaseDiscard,
			domain.MusPhaseGrande, domain.MusPhaseChica,
			domain.MusPhasePares, domain.MusPhaseJuego:
			if mi.Game.IsHumanTurn() {
				return
			}
			mi.Game.CpuPlay()
		case domain.MusPhaseShowdown:
			mi.Game.Showdown()
		case domain.MusPhaseRoundEnd:
			return
		default:
			return
		}
	}
}

// RestoreMusInteractor deserialises JSON into a MusInteractor.
func RestoreMusInteractor(data []byte, sp presenter.MusPresenter) (*MusInteractor, error) {
	return restoreAndBuild[domain.Mus](data, func(g *domain.Mus) *MusInteractor {
		return &MusInteractor{GameBase: GameBase[interfaces.MusGame]{Game: g}, sp: sp}
	})
}
