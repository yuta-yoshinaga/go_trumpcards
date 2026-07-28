//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// kempsMaxCpuSteps は CPU 自動進行の安全上限 (停止保証のフェイルセーフ)。
//
// フルCPU対戦を 1 回の advanceCpu 呼び出しで最後まで進める場合に備え、最悪
// ケース (KempsMaxRounds ラウンド × ラウンドあたりの最大ステップ数) を十分に
// 上回る値にしている。ドメイン側にもラウンド/スワップ上限があるため実際にこの
// 上限へ達することはない。
const kempsMaxCpuSteps = 8_000_000

// KempsInteractorIF はケムプスのインタラクターインタフェース。
type KempsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KempsConfig) string
	// Swap 人間が手札の 1 枚をフィールドの 1 枚と交換する
	Swap(handIndex, fieldIndex int) string
	// Pass 人間が交換せずにパスする
	Pass() string
	// SetSignal 人間が秘密のシグナル種別を設定する
	SetSignal(signalType int) string
	// DeclareKemps 人間が Kemps を宣言する
	DeclareKemps() string
	// DeclareCounterKemps 人間が相手 targetSeat に Counter-Kemps を宣言する
	DeclareCounterKemps(targetSeat int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KempsConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KempsInteractor はケムプスのインタラクタークラス。
type KempsInteractor struct {
	GameBase[interfaces.KempsGame]
	sp presenter.KempsPresenter
}

// NewKempsInteractor はコンストラクタ。
func NewKempsInteractor(g interfaces.KempsGame, sp presenter.KempsPresenter) *KempsInteractor {
	mustNotNil("KempsInteractor", map[string]any{"g": g, "sp": sp})
	return &KempsInteractor{GameBase: GameBase[interfaces.KempsGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (ki *KempsInteractor) Reset() string {
	ki.Game.Reset()
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KempsInteractor) ResetWithConfig(cfg domain.KempsConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.sp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Swap 人間が手札の 1 枚をフィールドの 1 枚と交換する
func (ki *KempsInteractor) Swap(handIndex, fieldIndex int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.sp); blocked {
		return out
	}
	if err := ki.Game.PlayerSwap(handIndex, fieldIndex); err != nil {
		return ki.sp.Output(ki.Game, err)
	}
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// Pass 人間が交換せずにパスする
func (ki *KempsInteractor) Pass() string {
	if out, blocked := guardNotPlayable(ki.Game, ki.sp); blocked {
		return out
	}
	if err := ki.Game.PlayerPass(); err != nil {
		return ki.sp.Output(ki.Game, err)
	}
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// SetSignal 人間が秘密のシグナル種別を設定する
func (ki *KempsInteractor) SetSignal(signalType int) string {
	if out, blocked := guardGameEnd(ki.Game, ki.sp); blocked {
		return out
	}
	ki.Game.PlayerSetSignal(signalType)
	return ki.sp.Output(ki.Game, nil)
}

// DeclareKemps 人間が Kemps を宣言する
func (ki *KempsInteractor) DeclareKemps() string {
	if out, blocked := guardGameEnd(ki.Game, ki.sp); blocked {
		return out
	}
	if err := ki.Game.PlayerDeclareKemps(); err != nil {
		return ki.sp.Output(ki.Game, err)
	}
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// DeclareCounterKemps 人間が相手 targetSeat に Counter-Kemps を宣言する
func (ki *KempsInteractor) DeclareCounterKemps(targetSeat int) string {
	if out, blocked := guardGameEnd(ki.Game, ki.sp); blocked {
		return out
	}
	if err := ki.Game.PlayerDeclareCounterKemps(targetSeat); err != nil {
		return ki.sp.Output(ki.Game, err)
	}
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ki *KempsInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ki.Game, ki.sp); blocked {
		return out
	}
	ki.Game.NextRound()
	ki.advanceCpu()
	return ki.sp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を取得
func (ki *KempsInteractor) GetConfig() domain.KempsConfig {
	return ki.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ki *KempsInteractor) ActionLog() string {
	return ki.sp.ActionLogOutput(ki.Game)
}

// advanceCpu は交換/宣言フェーズで人間の操作が必要になるまで CPU を自動進行
// させる。ラウンド終了 (KempsPhaseRoundEnd) に到達した場合は次のラウンドへ
// 自動的に進める (フルCPU対戦の連続進行に対応)。
func (ki *KempsInteractor) advanceCpu() {
	for step := 0; step < kempsMaxCpuSteps; step++ {
		if ki.Game.GetGameEndFlag() {
			return
		}
		switch ki.Game.GetPhase() {
		case domain.KempsPhaseExchange, domain.KempsPhaseDeclare:
			if ki.Game.IsHumanTurn() {
				return
			}
			ki.Game.CpuPlay()
		case domain.KempsPhaseRoundEnd:
			ki.Game.NextRound()
		default:
			return
		}
	}
}

// RestoreKempsInteractor deserialises JSON into a KempsInteractor.
func RestoreKempsInteractor(data []byte, sp presenter.KempsPresenter) (*KempsInteractor, error) {
	return restoreAndBuild[domain.Kemps](data, func(g *domain.Kemps) *KempsInteractor {
		return &KempsInteractor{GameBase: GameBase[interfaces.KempsGame]{Game: g}, sp: sp}
	})
}
