//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PigInteractorIF ピッグインタラクターインタフェース
type PigInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PigConfig) string
	// Pass 渡す札を選ぶ
	Pass(cardIndex int) string
	// Signal 合図に気づいたことを伝える
	Signal() string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PigConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PigInteractor ピッグインタラクタークラス
type PigInteractor struct {
	GameBase[interfaces.PigGame]
	sp presenter.PigPresenter
}

// NewPigInteractor コンストラクタ
func NewPigInteractor(s interfaces.PigGame, sp presenter.PigPresenter) *PigInteractor {
	mustNotNil("PigInteractor", map[string]any{"s": s, "sp": sp})
	return &PigInteractor{GameBase: GameBase[interfaces.PigGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化。人間の出番まで進める。
func (pi *PigInteractor) Reset() string {
	pi.Game.Reset()
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PigInteractor) ResetWithConfig(cfg domain.PigConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.sp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Pass 渡す札を選ぶ
func (pi *PigInteractor) Pass(cardIndex int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerPass(cardIndex); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// Signal 合図に気づいたことを伝える
func (pi *PigInteractor) Signal() string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.PlayerSignal(); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// NextRound 次のラウンドを配る
func (pi *PigInteractor) NextRound() string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	if err := pi.Game.NextRound(); err != nil {
		return pi.sp.Output(pi.Game, err)
	}
	pi.advance()
	return pi.sp.Output(pi.Game, nil)
}

// GiveUp 投了する
func (pi *PigInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(pi.Game, pi.sp); blocked {
		return out
	}
	pi.Game.GiveUp()
	return pi.sp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PigInteractor) GetConfig() domain.PigConfig { return pi.Game.GetConfig() }

// Hint ヒント取得
func (pi *PigInteractor) Hint() string { return pi.sp.HintOutput(pi.Game) }

// ActionLog 棋譜を出力する
func (pi *PigInteractor) ActionLog() string { return pi.sp.ActionLogOutput(pi.Game) }

// advance は人間の出番が来るまでゲームを進める。
//
// **合図の場面は放っておくと負けます。** CPU が順に気づき、最後に残った人が
// 文字を受け取る規則なので、ここで回し続けると人間が黙って負ける形になります。
// だからこそ人間が名乗れる間は必ず止まります（`IsHumanTurn` が真）。
func (pi *PigInteractor) advance() {
	for turns := 0; turns < maxCpuTurnsPerCall; turns++ {
		if pi.Game.GetGameEndFlag() || pi.Game.IsHumanTurn() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestorePigInteractor deserialises JSON into a PigInteractor.
func RestorePigInteractor(data []byte, sp presenter.PigPresenter) (*PigInteractor, error) {
	return restoreAndBuild[domain.Pig](data, func(g *domain.Pig) *PigInteractor {
		return NewPigInteractor(g, sp)
	})
}
