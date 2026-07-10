//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GoStopInteractorIF はゴーストップ (Go-Stop) のインタラクターインタフェース。
type GoStopInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.GoStopConfig) string
	// Play 手札を出す (fieldIdx で 2 枚一致時の捕獲対象を指定; 不要なら -1)
	Play(handIdx, fieldIdx int) string
	// Decide ゴー/ストップ決断 (true=ゴー, false=ストップ)
	Decide(goDecision bool) string
	// NextRound 次のラウンドを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.GoStopConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GoStopInteractor はゴーストップインタラクター。
type GoStopInteractor struct {
	GameBase[interfaces.GoStopGame]
	cp presenter.GoStopPresenter
}

// NewGoStopInteractor コンストラクタ。
func NewGoStopInteractor(kg interfaces.GoStopGame, cp presenter.GoStopPresenter) *GoStopInteractor {
	mustNotNil("GoStopInteractor", map[string]any{"kg": kg, "cp": cp})
	return &GoStopInteractor{GameBase: GameBase[interfaces.GoStopGame]{Game: kg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ki *GoStopInteractor) Reset() string {
	ki.Game.Reset()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ki *GoStopInteractor) ResetWithConfig(cfg domain.GoStopConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.cp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Play 手札を出す。
func (ki *GoStopInteractor) Play(handIdx, fieldIdx int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.cp); blocked {
		return out
	}
	if err := ki.Game.PlayerPlay(handIdx, fieldIdx); err != nil {
		return ki.cp.Output(ki.Game, err)
	}
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// Decide ゴー/ストップ決断。
func (ki *GoStopInteractor) Decide(goDecision bool) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.cp); blocked {
		return out
	}
	if err := ki.Game.PlayerDecide(goDecision); err != nil {
		return ki.cp.Output(ki.Game, err)
	}
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// NextRound 次のラウンドを開始する。
func (ki *GoStopInteractor) NextRound() string {
	ki.Game.NextRound()
	ki.advance()
	return ki.cp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ki *GoStopInteractor) GetConfig() domain.GoStopConfig { return ki.Game.GetConfig() }

// Hint ヒントを出力する。
func (ki *GoStopInteractor) Hint() string { return ki.cp.HintOutput(ki.Game) }

// ActionLog 棋譜を出力する。
func (ki *GoStopInteractor) ActionLog() string { return ki.cp.ActionLogOutput(ki.Game) }

// gostopMaxCpuIterations は advance の防御的な反復上限。
const gostopMaxCpuIterations = 1000

// advance はゲーム終了・ラウンド終了・人間の手番のいずれかに到達するまで CPU の
// プレイ/決断を回す。
func (ki *GoStopInteractor) advance() {
	for i := 0; i < gostopMaxCpuIterations; i++ {
		if ki.Game.GetGameEndFlag() {
			return
		}
		switch ki.Game.GetPhase() {
		case domain.GoStopPhasePlay:
			if ki.Game.IsHumanTurn() {
				return
			}
			ki.Game.CpuPlay()
		case domain.GoStopPhaseGoDecision:
			if ki.Game.IsHumanTurn() {
				return
			}
			ki.Game.CpuDecide()
		default:
			// RoundEnd / GameEnd は人間の操作待ち。
			return
		}
	}
}

// RestoreGoStopInteractor deserialises JSON into a GoStopInteractor.
func RestoreGoStopInteractor(data []byte, cp presenter.GoStopPresenter) (*GoStopInteractor, error) {
	return restoreAndBuild[domain.GoStop](data, func(g *domain.GoStop) *GoStopInteractor {
		return &GoStopInteractor{GameBase: GameBase[interfaces.GoStopGame]{Game: g}, cp: cp}
	})
}
