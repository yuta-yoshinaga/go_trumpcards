//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TeenDoPaanchInteractorIF 3-2-5 インタラクターインタフェース
type TeenDoPaanchInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TeenDoPaanchConfig) string
	// DeclareTrump 切り札を宣言する
	DeclareTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TeenDoPaanchConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TeenDoPaanchInteractor 3-2-5 インタラクタークラス
type TeenDoPaanchInteractor struct {
	GameBase[interfaces.TeenDoPaanchGame]
	tp presenter.TeenDoPaanchPresenter
}

// NewTeenDoPaanchInteractor コンストラクタ
func NewTeenDoPaanchInteractor(g interfaces.TeenDoPaanchGame, tp presenter.TeenDoPaanchPresenter) *TeenDoPaanchInteractor {
	mustNotNil("TeenDoPaanchInteractor", map[string]any{"g": g, "tp": tp})
	return &TeenDoPaanchInteractor{GameBase: GameBase[interfaces.TeenDoPaanchGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化。
//
// **切り札を決めるのはノルマ 5 の席。** 人間でなければ CPU に宣言させ、そのまま
// 人間の手番まで進める。宣言で止めると、人間が打てない盤面を返してしまう。
func (ti *TeenDoPaanchInteractor) Reset() string {
	ti.Game.Reset()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TeenDoPaanchInteractor) ResetWithConfig(cfg domain.TeenDoPaanchConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// DeclareTrump 切り札を宣言する
func (ti *TeenDoPaanchInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerDeclareTrump(suit); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TeenDoPaanchInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ti *TeenDoPaanchInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// GiveUp 投了する
func (ti *TeenDoPaanchInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.GiveUp()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TeenDoPaanchInteractor) GetConfig() domain.TeenDoPaanchConfig { return ti.Game.GetConfig() }

// Hint ヒント取得
func (ti *TeenDoPaanchInteractor) Hint() string { return ti.tp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する
func (ti *TeenDoPaanchInteractor) ActionLog() string { return ti.tp.ActionLogOutput(ti.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **宣言とプレイの両方を回す。** ノルマ 5 の席が CPU なら宣言させ、そのあと
// 人間の手番までカードを打たせる。ラウンド終了では止める（next は明示的に）。
func (ti *TeenDoPaanchInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if ti.Game.GetGameEndFlag() {
			return
		}
		switch ti.Game.GetPhase() {
		case domain.TeenDoPaanchPhaseTrump:
			if ti.Game.IsHumanTrumpTurn() {
				return
			}
			ti.Game.CpuDeclareTrump()
		case domain.TeenDoPaanchPhasePlay:
			if ti.Game.IsHumanTurn() {
				return
			}
			ti.Game.CpuPlay()
		default:
			return
		}
	}
}

// RestoreTeenDoPaanchInteractor deserialises JSON into a TeenDoPaanchInteractor.
func RestoreTeenDoPaanchInteractor(data []byte, tp presenter.TeenDoPaanchPresenter) (*TeenDoPaanchInteractor, error) {
	return restoreAndBuild[domain.TeenDoPaanch](data, func(g *domain.TeenDoPaanch) *TeenDoPaanchInteractor {
		return NewTeenDoPaanchInteractor(g, tp)
	})
}
