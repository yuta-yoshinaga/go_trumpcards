//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LiteratureInteractorIF リテラチャー (Literature) のインタラクターインタフェース
type LiteratureInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.LiteratureConfig) string
	// Ask 札を要求する
	Ask(to, suit, value int) string
	// Claim ハーフスートを宣言する
	Claim(half int, holders []int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.LiteratureConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LiteratureInteractor リテラチャー (Literature) のインタラクタークラス
type LiteratureInteractor struct {
	GameBase[interfaces.LiteratureGame]
	gp presenter.LiteraturePresenter
}

// NewLiteratureInteractor コンストラクタ
func NewLiteratureInteractor(g interfaces.LiteratureGame, gp presenter.LiteraturePresenter) *LiteratureInteractor {
	mustNotNil("LiteratureInteractor", map[string]any{"g": g, "gp": gp})
	return &LiteratureInteractor{GameBase: GameBase[interfaces.LiteratureGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (li *LiteratureInteractor) Reset() string {
	li.Game.Reset()
	li.runCpuTurns()
	return li.gp.Output(li.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (li *LiteratureInteractor) ResetWithConfig(cfg domain.LiteratureConfig) string {
	return resetWithValidatedConfig(li.Game, li.gp, cfg, li.Game.SetConfig, li.Reset)
}

// Ask 札を要求する
func (li *LiteratureInteractor) Ask(to, suit, value int) string {
	return li.act(func() error {
		return li.Game.Ask(li.Game.GetCurrentPlayerIdx(), to, domain.NewCard(suit, value, true))
	})
}

// Claim ハーフスートを宣言する
func (li *LiteratureInteractor) Claim(half int, holders []int) string {
	return li.act(func() error {
		return li.Game.Claim(li.Game.GetCurrentPlayerIdx(), half, holders)
	})
}

// act 人間アクションの共通処理 (ガード → 実行 → CPU 進行)
func (li *LiteratureInteractor) act(action func() error) string {
	if out, blocked := guardNotPlayable(li.Game, li.gp); blocked {
		return out
	}
	if err := action(); err != nil {
		return li.gp.Output(li.Game, err)
	}
	li.runCpuTurns()
	return li.gp.Output(li.Game, nil)
}

// GetConfig 現在の設定を取得
func (li *LiteratureInteractor) GetConfig() domain.LiteratureConfig {
	return li.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (li *LiteratureInteractor) ActionLog() string {
	return li.gp.ActionLogOutput(li.Game)
}

// literatureMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn or
// game end well within this limit).
const literatureMaxCpuSteps = 2000

// runCpuTurns CPUターンを連続実行する
func (li *LiteratureInteractor) runCpuTurns() {
	for step := 0; step < literatureMaxCpuSteps && !li.Game.GetGameEndFlag(); step++ {
		if li.Game.IsHumanTurn() {
			break
		}
		li.Game.CpuPlay()
	}
}

// RestoreLiteratureInteractor deserialises JSON into a LiteratureInteractor.
func RestoreLiteratureInteractor(data []byte, gp presenter.LiteraturePresenter) (*LiteratureInteractor, error) {
	return restoreAndBuild[domain.Literature](data, func(g *domain.Literature) *LiteratureInteractor {
		return &LiteratureInteractor{GameBase: GameBase[interfaces.LiteratureGame]{Game: g}, gp: gp}
	})
}
