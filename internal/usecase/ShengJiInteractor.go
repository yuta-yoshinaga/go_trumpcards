//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShengJiInteractorIF 升级 (Sheng Ji) のインタラクターインタフェース
type ShengJiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ShengJiConfig) string
	// Declare 亮牌する (ShengJiNoTrump はパス)
	Declare(suit int) string
	// BuryKitty 底牌に8枚埋め戻す
	BuryKitty(idxs []int) string
	// Play 手を出す
	Play(idxs []int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ShengJiConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ShengJiInteractor 升级 (Sheng Ji) のインタラクタークラス
type ShengJiInteractor struct {
	GameBase[interfaces.ShengJiGame]
	gp presenter.ShengJiPresenter
}

// NewShengJiInteractor コンストラクタ
func NewShengJiInteractor(g interfaces.ShengJiGame, gp presenter.ShengJiPresenter) *ShengJiInteractor {
	mustNotNil("ShengJiInteractor", map[string]any{"g": g, "gp": gp})
	return &ShengJiInteractor{GameBase: GameBase[interfaces.ShengJiGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (si *ShengJiInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *ShengJiInteractor) ResetWithConfig(cfg domain.ShengJiConfig) string {
	return resetWithValidatedConfig(si.Game, si.gp, cfg, si.Game.SetConfig, si.Reset)
}

// Declare 亮牌する
func (si *ShengJiInteractor) Declare(suit int) string {
	return si.act(func() error { return si.Game.Declare(si.Game.GetCurrentPlayerIdx(), suit) })
}

// BuryKitty 底牌に8枚埋め戻す
func (si *ShengJiInteractor) BuryKitty(idxs []int) string {
	return si.act(func() error { return si.Game.BuryKitty(si.Game.GetCurrentPlayerIdx(), idxs) })
}

// Play 手を出す
func (si *ShengJiInteractor) Play(idxs []int) string {
	return si.act(func() error { return si.Game.Play(si.Game.GetCurrentPlayerIdx(), idxs) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (si *ShengJiInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(si.Game, si.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return si.gp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// NextHand 次の局へ進む
func (si *ShengJiInteractor) NextHand() string {
	if out, blocked := guardGameEnd(si.Game, si.gp); blocked {
		return out
	}
	if err := si.Game.NextHand(); err != nil {
		return si.gp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.gp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *ShengJiInteractor) GetConfig() domain.ShengJiConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *ShengJiInteractor) ActionLog() string {
	return si.gp.ActionLogOutput(si.Game)
}

// shengJiMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever. **25 枚 × 4 人のトリックなので手数が多い**ぶん、
// 他のゲームより余裕を持たせてある。
const shengJiMaxCpuSteps = 2000

// runCpuTurns CPUターンを連続実行する
func (si *ShengJiInteractor) runCpuTurns() {
	for step := 0; step < shengJiMaxCpuSteps && !si.Game.GetGameEndFlag(); step++ {
		phase := si.Game.GetPhase()
		if phase == domain.ShengJiPhaseHandEnd || phase == domain.ShengJiPhaseGameEnd {
			break
		}
		if si.Game.IsHumanTurn() {
			break
		}
		si.Game.CpuPlay()
	}
}

// RestoreShengJiInteractor deserialises JSON into a ShengJiInteractor.
func RestoreShengJiInteractor(data []byte, gp presenter.ShengJiPresenter) (*ShengJiInteractor, error) {
	return restoreAndBuild[domain.ShengJi](data, func(g *domain.ShengJi) *ShengJiInteractor {
		return &ShengJiInteractor{GameBase: GameBase[interfaces.ShengJiGame]{Game: g}, gp: gp}
	})
}
