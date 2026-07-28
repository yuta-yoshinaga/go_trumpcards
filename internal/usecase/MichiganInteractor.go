//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MichiganInteractorIF はミシガン (Michigan) のインタラクターインタフェース。
type MichiganInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.MichiganConfig) string
	// Bet ブードル賭け (4 要素の分配) を適用する
	Bet(bets []int) string
	// Play 手札インデックスのカードを出す
	Play(cardIndex int) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.MichiganConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MichiganInteractor はミシガンのインタラクター。
type MichiganInteractor struct {
	GameBase[interfaces.MichiganGame]
	sp presenter.MichiganPresenter
}

// NewMichiganInteractor コンストラクタ。
func NewMichiganInteractor(g interfaces.MichiganGame, sp presenter.MichiganPresenter) *MichiganInteractor {
	mustNotNil("MichiganInteractor", map[string]any{"g": g, "sp": sp})
	return &MichiganInteractor{GameBase: GameBase[interfaces.MichiganGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *MichiganInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *MichiganInteractor) ResetWithConfig(cfg domain.MichiganConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bet ブードル賭けを適用する。
func (ti *MichiganInteractor) Bet(bets []int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.PlaceHumanBet(bets)
	})
}

// Play 手札インデックスのカードを出す。
func (ti *MichiganInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		return ti.Game.PlayCard(cardIndex)
	})
}

// NextRound 次のラウンドを配る。
func (ti *MichiganInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *MichiganInteractor) GetConfig() domain.MichiganConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *MichiganInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *MichiganInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestoreMichiganInteractor deserialises JSON into a MichiganInteractor.
func RestoreMichiganInteractor(data []byte, sp presenter.MichiganPresenter) (*MichiganInteractor, error) {
	return restoreAndBuild[domain.Michigan](data, func(g *domain.Michigan) *MichiganInteractor {
		return &MichiganInteractor{GameBase: GameBase[interfaces.MichiganGame]{Game: g}, sp: sp}
	})
}
