//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KarnoffelInteractorIF カルニッフェル (Karnöffel) のインタラクターインタフェース
type KarnoffelInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KarnoffelConfig) string
	// PlayCard 手札を1枚出す
	PlayCard(idx int) string
	// NextHand 次の局へ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KarnoffelConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KarnoffelInteractor カルニッフェル (Karnöffel) のインタラクタークラス
type KarnoffelInteractor struct {
	GameBase[interfaces.KarnoffelGame]
	gp presenter.KarnoffelPresenter
}

// NewKarnoffelInteractor コンストラクタ
func NewKarnoffelInteractor(g interfaces.KarnoffelGame, gp presenter.KarnoffelPresenter) *KarnoffelInteractor {
	mustNotNil("KarnoffelInteractor", map[string]any{"g": g, "gp": gp})
	return &KarnoffelInteractor{GameBase: GameBase[interfaces.KarnoffelGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ki *KarnoffelInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KarnoffelInteractor) ResetWithConfig(cfg domain.KarnoffelConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.gp, cfg, ki.Game.SetConfig, ki.Reset)
}

// PlayCard 手札を1枚出す
func (ki *KarnoffelInteractor) PlayCard(idx int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.gp); blocked {
		return out
	}
	if err := ki.Game.PlayCard(ki.Game.GetCurrentPlayerIdx(), idx); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// NextHand 次の局へ進む
func (ki *KarnoffelInteractor) NextHand() string {
	if out, blocked := guardGameEnd(ki.Game, ki.gp); blocked {
		return out
	}
	if err := ki.Game.NextHand(); err != nil {
		return ki.gp.Output(ki.Game, err)
	}
	ki.runCpuTurns()
	return ki.gp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を取得
func (ki *KarnoffelInteractor) GetConfig() domain.KarnoffelConfig {
	return ki.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ki *KarnoffelInteractor) ActionLog() string {
	return ki.gp.ActionLogOutput(ki.Game)
}

// karnoffelMaxCpuSteps bounds runCpuTurns so a malformed state can never spin
// the CPU loop forever (defensive — normal play always reaches a human turn,
// the settlement, or game end well within this limit).
const karnoffelMaxCpuSteps = 1000

// runCpuTurns CPUターンを連続実行する
func (ki *KarnoffelInteractor) runCpuTurns() {
	for step := 0; step < karnoffelMaxCpuSteps && !ki.Game.GetGameEndFlag(); step++ {
		phase := ki.Game.GetPhase()
		if phase == domain.KarnoffelPhaseHandEnd || phase == domain.KarnoffelPhaseGameEnd {
			break
		}
		if ki.Game.IsHumanTurn() {
			break
		}
		ki.Game.CpuPlay()
	}
}

// RestoreKarnoffelInteractor deserialises JSON into a KarnoffelInteractor.
func RestoreKarnoffelInteractor(data []byte, gp presenter.KarnoffelPresenter) (*KarnoffelInteractor, error) {
	return restoreAndBuild[domain.Karnoffel](data, func(g *domain.Karnoffel) *KarnoffelInteractor {
		return &KarnoffelInteractor{GameBase: GameBase[interfaces.KarnoffelGame]{Game: g}, gp: gp}
	})
}
