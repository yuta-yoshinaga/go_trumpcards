//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BouillotteInteractorIF はブイヨット (Bouillotte) のインタラクターインタフェース。
type BouillotteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.BouillotteConfig) string
	// Bet ベッティングアクション ("call"/"raise"/"fold") を適用する
	Bet(action string) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BouillotteConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BouillotteInteractor はブイヨットのインタラクター。
type BouillotteInteractor struct {
	GameBase[interfaces.BouillotteGame]
	sp presenter.BouillottePresenter
}

// NewBouillotteInteractor コンストラクタ。
func NewBouillotteInteractor(g interfaces.BouillotteGame, sp presenter.BouillottePresenter) *BouillotteInteractor {
	mustNotNil("BouillotteInteractor", map[string]any{"g": g, "sp": sp})
	return &BouillotteInteractor{GameBase: GameBase[interfaces.BouillotteGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ti *BouillotteInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.Reset)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ti *BouillotteInteractor) ResetWithConfig(cfg domain.BouillotteConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.sp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bet ベッティングアクションを適用する。
func (ti *BouillotteInteractor) Bet(action string) string {
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
func (ti *BouillotteInteractor) NextRound() string {
	return runAndPresent(ti.Game, ti.sp, ti.Game.NextRound)
}

// GetConfig 現在の設定を返す。
func (ti *BouillotteInteractor) GetConfig() domain.BouillotteConfig {
	return ti.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ti *BouillotteInteractor) Hint() string { return ti.sp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *BouillotteInteractor) ActionLog() string { return ti.sp.ActionLogOutput(ti.Game) }

// RestoreBouillotteInteractor deserialises JSON into a BouillotteInteractor.
func RestoreBouillotteInteractor(data []byte, sp presenter.BouillottePresenter) (*BouillotteInteractor, error) {
	return restoreAndBuild[domain.Bouillotte](data, func(g *domain.Bouillotte) *BouillotteInteractor {
		return &BouillotteInteractor{GameBase: GameBase[interfaces.BouillotteGame]{Game: g}, sp: sp}
	})
}
