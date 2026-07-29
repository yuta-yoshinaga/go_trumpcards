//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DuchessInteractorIF ダッチェス インタラクターインタフェース
type DuchessInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ChooseBaseRank 開始ランクを決める
	ChooseBaseRank(fanIdx int) string
	// Draw 山札からウェイストへ 1 枚めくる
	Draw() string
	// MoveReserveToFoundation リザーブから基礎札へ移動
	MoveReserveToFoundation(fanIdx int) string
	// MoveReserveToTableau リザーブからタブローへ移動
	MoveReserveToTableau(fanIdx, col int) string
	// MoveWasteToFoundation ウェイストから基礎札へ移動
	MoveWasteToFoundation() string
	// MoveWasteToTableau ウェイストからタブローへ移動
	MoveWasteToTableau(col int) string
	// MoveTableauToFoundation タブローから基礎札へ移動
	MoveTableauToFoundation(col int) string
	// MoveTableauToTableau タブロー間で移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
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

// DuchessInteractor ダッチェス インタラクタークラス
type DuchessInteractor struct {
	GameBase[interfaces.DuchessGame]
	dp presenter.DuchessPresenter
	solitaireActions[interfaces.DuchessGame]
}

// NewDuchessInteractor コンストラクタ
func NewDuchessInteractor(d interfaces.DuchessGame, dp presenter.DuchessPresenter) *DuchessInteractor {
	mustNotNil("DuchessInteractor", map[string]any{"d": d, "dp": dp})
	return &DuchessInteractor{
		GameBase:         GameBase[interfaces.DuchessGame]{Game: d},
		dp:               dp,
		solitaireActions: newSolitaireActions[interfaces.DuchessGame](d, dp),
	}
}

// Reset ゲーム初期化
func (di *DuchessInteractor) Reset() string {
	return runAndPresent(di.Game, di.dp, di.Game.Reset)
}

// ChooseBaseRank 開始ランクを決める
func (di *DuchessInteractor) ChooseBaseRank(fanIdx int) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.ChooseBaseRank(fanIdx) })
}

// Draw 山札からウェイストへ 1 枚めくる
func (di *DuchessInteractor) Draw() string {
	return execAndPresent(di.Game, di.dp, di.Game.Draw)
}

// MoveReserveToFoundation リザーブから基礎札へ移動
func (di *DuchessInteractor) MoveReserveToFoundation(fanIdx int) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.MoveReserveToFoundation(fanIdx) })
}

// MoveReserveToTableau リザーブからタブローへ移動
func (di *DuchessInteractor) MoveReserveToTableau(fanIdx, col int) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.MoveReserveToTableau(fanIdx, col) })
}

// MoveWasteToFoundation ウェイストから基礎札へ移動
func (di *DuchessInteractor) MoveWasteToFoundation() string {
	return execAndPresent(di.Game, di.dp, di.Game.MoveWasteToFoundation)
}

// MoveWasteToTableau ウェイストからタブローへ移動
func (di *DuchessInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.MoveWasteToTableau(col) })
}

// MoveTableauToFoundation タブローから基礎札へ移動
func (di *DuchessInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.MoveTableauToFoundation(col) })
}

// MoveTableauToTableau タブロー間で移動
func (di *DuchessInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(di.Game, di.dp, func() error {
		return di.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// Hint ヒント取得
func (di *DuchessInteractor) Hint() string {
	return di.dp.HintOutput(di.Game)
}

// ActionLog 棋譜を出力する
func (di *DuchessInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.Game)
}

// RestoreDuchessInteractor deserialises JSON into a DuchessInteractor.
func RestoreDuchessInteractor(data []byte, dp presenter.DuchessPresenter) (*DuchessInteractor, error) {
	return restoreAndBuild[domain.Duchess](data, func(g *domain.Duchess) *DuchessInteractor {
		return &DuchessInteractor{
			GameBase:         GameBase[interfaces.DuchessGame]{Game: g},
			dp:               dp,
			solitaireActions: newSolitaireActions[interfaces.DuchessGame](g, dp),
		}
	})
}
