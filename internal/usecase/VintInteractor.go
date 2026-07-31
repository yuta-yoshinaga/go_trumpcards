//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// VintInteractorIF ヴィント (Vint) のインタラクターインタフェース
type VintInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.VintConfig) string
	// Bid 宣言する
	Bid(level, denom int) string
	// PassBid 宣言を見送る
	PassBid() string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.VintConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// VintInteractor ヴィント (Vint) のインタラクタークラス
type VintInteractor struct {
	GameBase[interfaces.VintGame]
	gp presenter.VintPresenter
}

// NewVintInteractor コンストラクタ
func NewVintInteractor(g interfaces.VintGame, gp presenter.VintPresenter) *VintInteractor {
	mustNotNil("VintInteractor", map[string]any{"g": g, "gp": gp})
	return &VintInteractor{GameBase: GameBase[interfaces.VintGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (vi *VintInteractor) Reset() string {
	vi.Game.Reset()
	vi.runCpuTurns()
	return vi.gp.Output(vi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (vi *VintInteractor) ResetWithConfig(cfg domain.VintConfig) string {
	return resetWithValidatedConfig(vi.Game, vi.gp, cfg, vi.Game.SetConfig, vi.Reset)
}

// Bid 宣言する
func (vi *VintInteractor) Bid(level, denom int) string {
	return vi.act(func() error { return vi.Game.Bid(vi.Game.GetBidPlayerIdx(), level, denom) })
}

// PassBid 宣言を見送る
func (vi *VintInteractor) PassBid() string {
	return vi.act(func() error { return vi.Game.PassBid(vi.Game.GetBidPlayerIdx()) })
}

// PlayCard 手札を1枚出す
func (vi *VintInteractor) PlayCard(idx int) string {
	return vi.act(func() error { return vi.Game.PlayCard(vi.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (vi *VintInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(vi.Game, vi.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return vi.gp.Output(vi.Game, err)
	}
	vi.runCpuTurns()
	return vi.gp.Output(vi.Game, nil)
}

// NextHand 次の局へ進む
func (vi *VintInteractor) NextHand() string {
	if out, blocked := guardGameEnd(vi.Game, vi.gp); blocked {
		return out
	}
	if err := vi.Game.NextHand(); err != nil {
		return vi.gp.Output(vi.Game, err)
	}
	vi.runCpuTurns()
	return vi.gp.Output(vi.Game, nil)
}

// GetConfig 現在の設定を取得
func (vi *VintInteractor) GetConfig() domain.VintConfig {
	return vi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (vi *VintInteractor) ActionLog() string {
	return vi.gp.ActionLogOutput(vi.Game)
}

// vintMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever (defensive — normal play always reaches a human turn, the
// settlement, or game end well within this limit).
const vintMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (vi *VintInteractor) runCpuTurns() {
	for step := 0; step < vintMaxCpuSteps && !vi.Game.GetGameEndFlag(); step++ {
		phase := vi.Game.GetPhase()
		if phase == domain.VintPhaseHandEnd || phase == domain.VintPhaseGameEnd {
			break
		}
		if vi.Game.IsHumanTurn() {
			break
		}
		vi.Game.CpuPlay()
	}
}

// RestoreVintInteractor deserialises JSON into a VintInteractor.
func RestoreVintInteractor(data []byte, gp presenter.VintPresenter) (*VintInteractor, error) {
	return restoreAndBuild[domain.Vint](data, func(g *domain.Vint) *VintInteractor {
		return &VintInteractor{GameBase: GameBase[interfaces.VintGame]{Game: g}, gp: gp}
	})
}
