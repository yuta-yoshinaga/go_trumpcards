//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GrandfathersClockInteractorIF グランドファーザーズ・クロック インタラクターインタフェース
type GrandfathersClockInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToFoundation タブローから文字盤へ移動
	MoveTableauToFoundation(col, fIdx int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, toCol int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// GrandfathersClockInteractor グランドファーザーズ・クロック インタラクタークラス
type GrandfathersClockInteractor struct {
	GameBase[interfaces.GrandfathersClockGame]
	gcp presenter.GrandfathersClockPresenter
	solitaireActions[interfaces.GrandfathersClockGame]
}

// NewGrandfathersClockInteractor コンストラクタ
func NewGrandfathersClockInteractor(gc interfaces.GrandfathersClockGame, gcp presenter.GrandfathersClockPresenter) *GrandfathersClockInteractor {
	mustNotNil("GrandfathersClockInteractor", map[string]any{"gc": gc, "gcp": gcp})
	return &GrandfathersClockInteractor{
		GameBase:         GameBase[interfaces.GrandfathersClockGame]{Game: gc},
		gcp:              gcp,
		solitaireActions: newSolitaireActions[interfaces.GrandfathersClockGame](gc, gcp),
	}
}

// Reset ゲーム初期化
func (gi *GrandfathersClockInteractor) Reset() string {
	return runAndPresent(gi.Game, gi.gcp, gi.Game.Reset)
}

// MoveTableauToFoundation タブローから文字盤へ移動
func (gi *GrandfathersClockInteractor) MoveTableauToFoundation(col, fIdx int) string {
	return execAndPresent(gi.Game, gi.gcp, func() error { return gi.Game.MoveTableauToFoundation(col, fIdx) })
}

// MoveTableauToTableau タブロー間で移動
func (gi *GrandfathersClockInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(gi.Game, gi.gcp, func() error { return gi.Game.MoveTableauToTableau(fromCol, toCol) })
}

// Hint ヒント取得
func (gi *GrandfathersClockInteractor) Hint() string {
	return gi.gcp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *GrandfathersClockInteractor) ActionLog() string {
	return gi.gcp.ActionLogOutput(gi.Game)
}

// RestoreGrandfathersClockInteractor deserialises JSON into a GrandfathersClockInteractor.
func RestoreGrandfathersClockInteractor(data []byte, gcp presenter.GrandfathersClockPresenter) (*GrandfathersClockInteractor, error) {
	return restoreAndBuild[domain.GrandfathersClock](data, func(g *domain.GrandfathersClock) *GrandfathersClockInteractor {
		return &GrandfathersClockInteractor{
			GameBase:         GameBase[interfaces.GrandfathersClockGame]{Game: g},
			gcp:              gcp,
			solitaireActions: newSolitaireActions[interfaces.GrandfathersClockGame](g, gcp),
		}
	})
}
