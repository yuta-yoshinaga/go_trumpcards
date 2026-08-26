//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BigBenInteractorIF ビッグ・ベン インタラクターインタフェース
type BigBenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToFoundation タブローから文字盤へ移動
	MoveTableauToFoundation(col, fIdx int) string
	// Deal 山札から補充する
	Deal() string
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

// BigBenInteractor ビッグ・ベン インタラクタークラス
type BigBenInteractor struct {
	GameBase[interfaces.BigBenGame]
	gcp presenter.BigBenPresenter
	solitaireActions[interfaces.BigBenGame]
}

// NewBigBenInteractor コンストラクタ
func NewBigBenInteractor(gc interfaces.BigBenGame, gcp presenter.BigBenPresenter) *BigBenInteractor {
	mustNotNil("BigBenInteractor", map[string]any{"gc": gc, "gcp": gcp})
	return &BigBenInteractor{
		GameBase:         GameBase[interfaces.BigBenGame]{Game: gc},
		gcp:              gcp,
		solitaireActions: newSolitaireActions[interfaces.BigBenGame](gc, gcp),
	}
}

// Reset ゲーム初期化
func (gi *BigBenInteractor) Reset() string {
	return runAndPresent(gi.Game, gi.gcp, gi.Game.Reset)
}

// Deal 山札から補充する
func (gi *BigBenInteractor) Deal() string {
	return execAndPresent(gi.Game, gi.gcp, gi.Game.Deal)
}

// MoveTableauToFoundation タブローから文字盤へ移動
func (gi *BigBenInteractor) MoveTableauToFoundation(col, fIdx int) string {
	return execAndPresent(gi.Game, gi.gcp, func() error { return gi.Game.MoveTableauToFoundation(col, fIdx) })
}

// MoveTableauToTableau タブロー間で移動
func (gi *BigBenInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(gi.Game, gi.gcp, func() error { return gi.Game.MoveTableauToTableau(fromCol, toCol) })
}

// Hint ヒント取得
func (gi *BigBenInteractor) Hint() string {
	return gi.gcp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *BigBenInteractor) ActionLog() string {
	return gi.gcp.ActionLogOutput(gi.Game)
}

// RestoreBigBenInteractor deserialises JSON into a BigBenInteractor.
func RestoreBigBenInteractor(data []byte, gcp presenter.BigBenPresenter) (*BigBenInteractor, error) {
	return restoreAndBuild[domain.BigBen](data, func(g *domain.BigBen) *BigBenInteractor {
		return &BigBenInteractor{
			GameBase:         GameBase[interfaces.BigBenGame]{Game: g},
			gcp:              gcp,
			solitaireActions: newSolitaireActions[interfaces.BigBenGame](g, gcp),
		}
	})
}
