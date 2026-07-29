//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WindmillInteractorIF ウィンドミル インタラクターインタフェース
type WindmillInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から捨て札へ 1 枚めくる
	Draw() string
	// MoveSailToCenter 帆から中央基礎札へ移動
	MoveSailToCenter(sailIdx int) string
	// MoveSailToCorner 帆から四隅の基礎札へ移動
	MoveSailToCorner(sailIdx, cornerIdx int) string
	// MoveWasteToCenter 捨て札から中央基礎札へ移動
	MoveWasteToCenter() string
	// MoveWasteToCorner 捨て札から四隅の基礎札へ移動
	MoveWasteToCorner(cornerIdx int) string
	// MoveCornerToCenter 四隅から中央基礎札へ引き戻す
	MoveCornerToCenter(cornerIdx int) string
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

// WindmillInteractor ウィンドミル インタラクタークラス
type WindmillInteractor struct {
	GameBase[interfaces.WindmillGame]
	wp presenter.WindmillPresenter
	solitaireActions[interfaces.WindmillGame]
}

// NewWindmillInteractor コンストラクタ
func NewWindmillInteractor(w interfaces.WindmillGame, wp presenter.WindmillPresenter) *WindmillInteractor {
	mustNotNil("WindmillInteractor", map[string]any{"w": w, "wp": wp})
	return &WindmillInteractor{
		GameBase:         GameBase[interfaces.WindmillGame]{Game: w},
		wp:               wp,
		solitaireActions: newSolitaireActions[interfaces.WindmillGame](w, wp),
	}
}

// Reset ゲーム初期化
func (wi *WindmillInteractor) Reset() string {
	return runAndPresent(wi.Game, wi.wp, wi.Game.Reset)
}

// Draw 山札から捨て札へ 1 枚めくる
func (wi *WindmillInteractor) Draw() string {
	return execAndPresent(wi.Game, wi.wp, wi.Game.Draw)
}

// MoveSailToCenter 帆から中央基礎札へ移動
func (wi *WindmillInteractor) MoveSailToCenter(sailIdx int) string {
	return execAndPresent(wi.Game, wi.wp, func() error { return wi.Game.MoveSailToCenter(sailIdx) })
}

// MoveSailToCorner 帆から四隅の基礎札へ移動
func (wi *WindmillInteractor) MoveSailToCorner(sailIdx, cornerIdx int) string {
	return execAndPresent(wi.Game, wi.wp, func() error { return wi.Game.MoveSailToCorner(sailIdx, cornerIdx) })
}

// MoveWasteToCenter 捨て札から中央基礎札へ移動
func (wi *WindmillInteractor) MoveWasteToCenter() string {
	return execAndPresent(wi.Game, wi.wp, wi.Game.MoveWasteToCenter)
}

// MoveWasteToCorner 捨て札から四隅の基礎札へ移動
func (wi *WindmillInteractor) MoveWasteToCorner(cornerIdx int) string {
	return execAndPresent(wi.Game, wi.wp, func() error { return wi.Game.MoveWasteToCorner(cornerIdx) })
}

// MoveCornerToCenter 四隅から中央基礎札へ引き戻す
func (wi *WindmillInteractor) MoveCornerToCenter(cornerIdx int) string {
	return execAndPresent(wi.Game, wi.wp, func() error { return wi.Game.MoveCornerToCenter(cornerIdx) })
}

// Hint ヒント取得
func (wi *WindmillInteractor) Hint() string {
	return wi.wp.HintOutput(wi.Game)
}

// ActionLog 棋譜を出力する
func (wi *WindmillInteractor) ActionLog() string {
	return wi.wp.ActionLogOutput(wi.Game)
}

// RestoreWindmillInteractor deserialises JSON into a WindmillInteractor.
func RestoreWindmillInteractor(data []byte, wp presenter.WindmillPresenter) (*WindmillInteractor, error) {
	return restoreAndBuild[domain.Windmill](data, func(g *domain.Windmill) *WindmillInteractor {
		return &WindmillInteractor{
			GameBase:         GameBase[interfaces.WindmillGame]{Game: g},
			wp:               wp,
			solitaireActions: newSolitaireActions[interfaces.WindmillGame](g, wp),
		}
	})
}
