//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ScorpionInteractorIF スコーピオンインタラクターインタフェース
type ScorpionInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Deal ストックからタブローに配る
	Deal() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// LegalMoves 指定列のトップカードの合法な移動先を出力する
	LegalMoves(col int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
	// UndoToEscape 膠着状態脱出に必要なアンドゥ手数を返す
	UndoToEscape() int
}

// ScorpionInteractor スコーピオンインタラクタークラス
type ScorpionInteractor struct {
	GameBase[interfaces.ScorpionGame]
	sp presenter.ScorpionPresenter
	solitaireActions[interfaces.ScorpionGame]
}

// NewScorpionInteractor コンストラクタ
func NewScorpionInteractor(s interfaces.ScorpionGame, sp presenter.ScorpionPresenter) *ScorpionInteractor {
	mustNotNil("ScorpionInteractor", map[string]any{"s": s, "sp": sp})
	return &ScorpionInteractor{
		GameBase:         GameBase[interfaces.ScorpionGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.ScorpionGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *ScorpionInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Deal ストックからタブローに配る
func (si *ScorpionInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.Deal)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (si *ScorpionInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error {
		return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// Hint ヒント取得
func (si *ScorpionInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// LegalMoves 指定列のトップカードの合法な移動先を出力する
func (si *ScorpionInteractor) LegalMoves(col int) string {
	return si.sp.LegalMovesOutput(si.Game, col)
}

// ActionLog 棋譜を出力する
func (si *ScorpionInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// UndoToEscape 膠着状態脱出に必要なアンドゥ手数を返す
func (si *ScorpionInteractor) UndoToEscape() int {
	return si.Game.UndoToEscape()
}

// RestoreScorpionInteractor deserialises JSON into a ScorpionInteractor.
func RestoreScorpionInteractor(data []byte, sp presenter.ScorpionPresenter) (*ScorpionInteractor, error) {
	return restoreAndBuild[domain.Scorpion](data, func(g *domain.Scorpion) *ScorpionInteractor {
		return &ScorpionInteractor{
			GameBase:         GameBase[interfaces.ScorpionGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.ScorpionGame](g, sp),
		}
	})
}
