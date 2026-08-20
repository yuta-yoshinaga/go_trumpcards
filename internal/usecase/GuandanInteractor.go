//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GuandanInteractorIF 掼蛋 (Guandan) のインタラクターインタフェース
type GuandanInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GuandanConfig) string
	// PlayCards 手札から役を出す
	PlayCards(idxs []int) string
	// Pass パスする
	Pass() string
	// ReturnTribute 還貢する
	ReturnTribute(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GuandanConfig
	// Check 手札の組み合わせが何の役になるかを調べる
	Check(idxs []int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GuandanInteractor 掼蛋 (Guandan) のインタラクタークラス
type GuandanInteractor struct {
	GameBase[interfaces.GuandanGame]
	gp presenter.GuandanPresenter
}

// NewGuandanInteractor コンストラクタ
func NewGuandanInteractor(g interfaces.GuandanGame, gp presenter.GuandanPresenter) *GuandanInteractor {
	mustNotNil("GuandanInteractor", map[string]any{"g": g, "gp": gp})
	return &GuandanInteractor{GameBase: GameBase[interfaces.GuandanGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (gi *GuandanInteractor) Reset() string {
	gi.Game.Reset()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (gi *GuandanInteractor) ResetWithConfig(cfg domain.GuandanConfig) string {
	return resetWithValidatedConfig(gi.Game, gi.gp, cfg, gi.Game.SetConfig, gi.Reset)
}

// PlayCards 手札から役を出す
func (gi *GuandanInteractor) PlayCards(idxs []int) string {
	return gi.act(func() error { return gi.Game.PlayCards(gi.Game.GetCurrentPlayerIdx(), idxs) })
}

// Pass パスする
func (gi *GuandanInteractor) Pass() string {
	return gi.act(func() error { return gi.Game.Pass(gi.Game.GetCurrentPlayerIdx()) })
}

// ReturnTribute 還貢する
func (gi *GuandanInteractor) ReturnTribute(idx int) string {
	return gi.act(func() error { return gi.Game.ReturnTribute(gi.Game.GetCurrentPlayerIdx(), idx) })
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (gi *GuandanInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// NextHand 次の局へ進む
func (gi *GuandanInteractor) NextHand() string {
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	if err := gi.Game.NextHand(); err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// GetConfig 現在の設定を取得
func (gi *GuandanInteractor) GetConfig() domain.GuandanConfig {
	return gi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
// Check 手札の組み合わせが何の役になるかを調べる
func (gi *GuandanInteractor) Check(idxs []int) string { return gi.gp.CheckOutput(gi.Game, idxs) }

func (gi *GuandanInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// guandanMaxCpuSteps bounds runCpuTurns so a malformed state can never spin the
// CPU loop forever. **27 枚 × 4 人のクライミングなので手数が多い**ぶん、
// 他のゲームより余裕を持たせてある。
const guandanMaxCpuSteps = 5000

// runCpuTurns CPUターンを連続実行する
func (gi *GuandanInteractor) runCpuTurns() {
	for step := 0; step < guandanMaxCpuSteps && !gi.Game.GetGameEndFlag(); step++ {
		phase := gi.Game.GetPhase()
		if phase == domain.GuandanPhaseHandEnd || phase == domain.GuandanPhaseGameEnd {
			break
		}
		if gi.Game.IsHumanTurn() {
			break
		}
		gi.Game.CpuPlay()
	}
}

// RestoreGuandanInteractor deserialises JSON into a GuandanInteractor.
func RestoreGuandanInteractor(data []byte, gp presenter.GuandanPresenter) (*GuandanInteractor, error) {
	return restoreAndBuild[domain.Guandan](data, func(g *domain.Guandan) *GuandanInteractor {
		return &GuandanInteractor{GameBase: GameBase[interfaces.GuandanGame]{Game: g}, gp: gp}
	})
}
