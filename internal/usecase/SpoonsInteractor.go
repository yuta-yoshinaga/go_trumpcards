//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// spoonsMaxCpuSteps は CPU 自動進行の安全上限 (停止保証のフェイルセーフ)。
//
// フルCPU対戦を 1 回の advanceCpu 呼び出しで最後まで進める場合に備え、最悪
// ケース (SpoonsMaxRounds ラウンド × ラウンドあたりの最大ステップ数) を十分に
// 上回る値にしている。ドメイン側にもラウンド/パス上限があるため実際にこの上限へ
// 達することはない。
const spoonsMaxCpuSteps = 2_000_000

// SpoonsInteractorIF はスプーンのインタラクターインタフェース。
type SpoonsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SpoonsConfig) string
	// Pass 人間が手札の 1 枚を次へ渡す
	Pass(cardIndex int) string
	// Grab 人間がスプーンを掴む
	Grab() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SpoonsConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpoonsInteractor はスプーンのインタラクタークラス。
type SpoonsInteractor struct {
	GameBase[interfaces.SpoonsGame]
	sp presenter.SpoonsPresenter
}

// NewSpoonsInteractor はコンストラクタ。
func NewSpoonsInteractor(g interfaces.SpoonsGame, sp presenter.SpoonsPresenter) *SpoonsInteractor {
	mustNotNil("SpoonsInteractor", map[string]any{"g": g, "sp": sp})
	return &SpoonsInteractor{GameBase: GameBase[interfaces.SpoonsGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SpoonsInteractor) Reset() string {
	si.Game.Reset()
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SpoonsInteractor) ResetWithConfig(cfg domain.SpoonsConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Pass 人間が手札の 1 枚を次へ渡す
func (si *SpoonsInteractor) Pass(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPass(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// Grab 人間がスプーンを掴む
func (si *SpoonsInteractor) Grab() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerGrabSpoon(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンドへ進む
func (si *SpoonsInteractor) NextRound() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.advanceCpu()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SpoonsInteractor) GetConfig() domain.SpoonsConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SpoonsInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// advanceCpu はパス/グラブフェーズで人間の操作が必要になるまで CPU を自動進行
// させる。ラウンド終了 (SpoonsPhaseRoundEnd) に到達した場合は次のラウンドへ
// 自動的に進める (フルCPU対戦の連続進行に対応)。
func (si *SpoonsInteractor) advanceCpu() {
	for step := 0; step < spoonsMaxCpuSteps; step++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		switch si.Game.GetPhase() {
		case domain.SpoonsPhasePass, domain.SpoonsPhaseGrab:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
		case domain.SpoonsPhaseRoundEnd:
			si.Game.NextRound()
		default:
			return
		}
	}
}

// RestoreSpoonsInteractor deserialises JSON into a SpoonsInteractor.
func RestoreSpoonsInteractor(data []byte, sp presenter.SpoonsPresenter) (*SpoonsInteractor, error) {
	return restoreAndBuild[domain.Spoons](data, func(g *domain.Spoons) *SpoonsInteractor {
		return &SpoonsInteractor{GameBase: GameBase[interfaces.SpoonsGame]{Game: g}, sp: sp}
	})
}
