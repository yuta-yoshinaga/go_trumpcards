package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// YukonInteractorIF ユーコンインタラクターインタフェース
type YukonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
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

// YukonInteractor ユーコンインタラクタークラス
type YukonInteractor struct {
	GameBase[interfaces.YukonGame]
	yp presenter.YukonPresenter
	solitaireActions[interfaces.YukonGame]
}

// NewYukonInteractor コンストラクタ
func NewYukonInteractor(y interfaces.YukonGame, yp presenter.YukonPresenter) *YukonInteractor {
	mustNotNil("YukonInteractor", map[string]any{"y": y, "yp": yp})
	return &YukonInteractor{
		GameBase:         GameBase[interfaces.YukonGame]{Game: y},
		yp:               yp,
		solitaireActions: newSolitaireActions[interfaces.YukonGame](y, yp),
	}
}

// Reset ゲーム初期化
func (yi *YukonInteractor) Reset() string {
	return runAndPresent(yi.Game, yi.yp, yi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (yi *YukonInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(yi.Game, yi.yp, func() error {
		return yi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol)
	})
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (yi *YukonInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(yi.Game, yi.yp, func() error { return yi.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (yi *YukonInteractor) Hint() string {
	return yi.yp.HintOutput(yi.Game)
}

// ActionLog 棋譜を出力する
func (yi *YukonInteractor) ActionLog() string {
	return yi.yp.ActionLogOutput(yi.Game)
}

// RestoreYukonInteractor deserialises JSON into a YukonInteractor.
func RestoreYukonInteractor(data []byte, yp presenter.YukonPresenter) (*YukonInteractor, error) {
	return restoreAndBuild[domain.Yukon](data, func(g *domain.Yukon) *YukonInteractor {
		return &YukonInteractor{
			GameBase:         GameBase[interfaces.YukonGame]{Game: g},
			yp:               yp,
			solitaireActions: newSolitaireActions[interfaces.YukonGame](g, yp),
		}
	})
}
