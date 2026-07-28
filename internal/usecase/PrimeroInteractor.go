//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PrimeroInteractorIF はプリメロ (Primero) のインタラクターインタフェース。
type PrimeroInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.PrimeroConfig) string
	// Bet ベッティングアクション ("call"/"raise"/"fold") を適用する
	Bet(action string) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.PrimeroConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PrimeroInteractor はプリメロのインタラクター。
type PrimeroInteractor struct {
	GameBase[interfaces.PrimeroGame]
	sp presenter.PrimeroPresenter
}

// NewPrimeroInteractor コンストラクタ。
func NewPrimeroInteractor(g interfaces.PrimeroGame, sp presenter.PrimeroPresenter) *PrimeroInteractor {
	mustNotNil("PrimeroInteractor", map[string]any{"g": g, "sp": sp})
	return &PrimeroInteractor{GameBase: GameBase[interfaces.PrimeroGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *PrimeroInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *PrimeroInteractor) ResetWithConfig(cfg domain.PrimeroConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bet ベッティングアクションを適用する。
func (ti *PrimeroInteractor) Bet(action string) string {
	if out, blocked := guardGameEnd(ti.Game, ti.sp); blocked {
		return out
	}
	return execAndPresent(ti.Game, ti.sp, func() error {
		switch action {
		case "call":
			return ti.Game.PlayerCall()
		case "raise":
			return ti.Game.PlayerRaise()
		case "fold":
			return ti.Game.PlayerFold()
		default:
			return domain.NewDomainError(domain.ErrInvalidPlay, "unknown betting action")
		}
	})
}

// NextRound 次のラウンドを配る。
func (ti *PrimeroInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *PrimeroInteractor) GetConfig() domain.PrimeroConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *PrimeroInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *PrimeroInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestorePrimeroInteractor deserialises JSON into a PrimeroInteractor.
func RestorePrimeroInteractor(data []byte, sp presenter.PrimeroPresenter) (*PrimeroInteractor, error) {
	return restoreAndBuild[domain.Primero](data, func(g *domain.Primero) *PrimeroInteractor {
		return &PrimeroInteractor{GameBase: GameBase[interfaces.PrimeroGame]{Game: g}, sp: sp}
	})
}
