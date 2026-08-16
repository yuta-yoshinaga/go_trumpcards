//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SambaInteractorIF サンバインタラクターインタフェース
type SambaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SambaConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// DrawFromDiscard 捨て札の山を取る
	DrawFromDiscard(naturalPairIndices []int) string
	// Meld メルドを出す
	Meld(meldGroups [][]int) string
	// SkipMeld メルドフェーズをスキップ
	SkipMeld() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// GoOut 上がる
	GoOut() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SambaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SambaInteractor サンバインタラクタークラス
type SambaInteractor struct {
	GameBase[interfaces.SambaGame]
	gp presenter.SambaPresenter
}

// NewSambaInteractor コンストラクタ
func NewSambaInteractor(g interfaces.SambaGame, gp presenter.SambaPresenter) *SambaInteractor {
	mustNotNil("SambaInteractor", map[string]any{"g": g, "gp": gp})
	return &SambaInteractor{GameBase: GameBase[interfaces.SambaGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *SambaInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *SambaInteractor) ResetWithConfig(cfg domain.SambaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *SambaInteractor) DrawFromStock() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromStock()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// DrawFromDiscard 捨て札の山を取る
func (ci *SambaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDrawFromDiscard(naturalPairIndices)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Meld メルドを出す
func (ci *SambaInteractor) Meld(meldGroups [][]int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerMeld(meldGroups)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// SkipMeld メルドフェーズをスキップ
func (ci *SambaInteractor) SkipMeld() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerSkipMeld()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Discard カードを捨てる
func (ci *SambaInteractor) Discard(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDiscard(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GoOut 上がる
func (ci *SambaInteractor) GoOut() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerGoOut()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ci *SambaInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *SambaInteractor) GetConfig() domain.SambaConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *SambaInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// runCpuTurns CPUターンを実行
func (ci *SambaInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.SambaPhaseRoundEnd || phase == domain.SambaPhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreSambaInteractor deserialises JSON into a SambaInteractor.
func RestoreSambaInteractor(data []byte, gp presenter.SambaPresenter) (*SambaInteractor, error) {
	return restoreAndBuild[domain.Samba](data, func(g *domain.Samba) *SambaInteractor {
		return &SambaInteractor{GameBase: GameBase[interfaces.SambaGame]{Game: g}, gp: gp}
	})
}
