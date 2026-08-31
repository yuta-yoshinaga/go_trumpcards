//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BristolInteractorIF ブリストルインタラクターインタフェース
type BristolInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックから3つのファンに1枚ずつ配る
	Draw() string
	// MoveTableauToTableau タブローからタブローに移動
	MoveTableauToTableau(fromCol, toCol int) string
	// MoveTableauToFoundation タブローからファウンデーションに移動
	MoveTableauToFoundation(col int) string
	// MoveFanToTableau ファンからタブローに移動
	MoveFanToTableau(fanIdx, toCol int) string
	// MoveFanToFoundation ファンからファウンデーションに移動
	MoveFanToFoundation(fanIdx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// Targets 移動元ゾーン zone の列 col の札を置ける先を一覧する
	Targets(zone string, col int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜出力
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// BristolInteractor ブリストルインタラクタークラス
type BristolInteractor struct {
	GameBase[interfaces.BristolGame]
	op presenter.BristolPresenter
}

// NewBristolInteractor コンストラクタ
func NewBristolInteractor(b interfaces.BristolGame, op presenter.BristolPresenter) *BristolInteractor {
	mustNotNil("BristolInteractor", map[string]any{"b": b, "op": op})
	return &BristolInteractor{GameBase: GameBase[interfaces.BristolGame]{Game: b}, op: op}
}

// Reset ゲーム初期化
func (bi *BristolInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.op, bi.Game.Reset)
}

// Draw ストックから3つのファンに1枚ずつ配る
func (bi *BristolInteractor) Draw() string {
	return execAndPresent(bi.Game, bi.op, bi.Game.Draw)
}

// MoveTableauToTableau タブローからタブローに移動
func (bi *BristolInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return execAndPresent(bi.Game, bi.op, func() error { return bi.Game.MoveTableauToTableau(fromCol, toCol) })
}

// MoveTableauToFoundation タブローからファウンデーションに移動
func (bi *BristolInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.op, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// MoveFanToTableau ファンからタブローに移動
func (bi *BristolInteractor) MoveFanToTableau(fanIdx, toCol int) string {
	return execAndPresent(bi.Game, bi.op, func() error { return bi.Game.MoveFanToTableau(fanIdx, toCol) })
}

// MoveFanToFoundation ファンからファウンデーションに移動
func (bi *BristolInteractor) MoveFanToFoundation(fanIdx int) string {
	return execAndPresent(bi.Game, bi.op, func() error { return bi.Game.MoveFanToFoundation(fanIdx) })
}

// GiveUp ギブアップ
func (bi *BristolInteractor) GiveUp() string {
	return runAndPresent(bi.Game, bi.op, bi.Game.GiveUp)
}

// Hint ヒント取得
func (bi *BristolInteractor) Hint() string {
	return bi.op.HintOutput(bi.Game)
}

// Targets 移動元ゾーン zone の列 col の札を置ける先を一覧する
func (bi *BristolInteractor) Targets(zone string, col int) string {
	return bi.op.TargetsOutput(bi.Game, zone, col)
}

// AutoComplete オートコンプリート
func (bi *BristolInteractor) AutoComplete() string {
	return execAndPresent(bi.Game, bi.op, bi.Game.AutoComplete)
}

// ActionLog 棋譜出力
func (bi *BristolInteractor) ActionLog() string {
	return bi.op.ActionLogOutput(bi.Game)
}

// Undo アンドゥ
func (bi *BristolInteractor) Undo() string {
	return execAndPresent(bi.Game, bi.op, bi.Game.Undo)
}

// UndoN n回連続アンドゥ
func (bi *BristolInteractor) UndoN(n int) string {
	return execAndPresent(bi.Game, bi.op, func() error { return bi.Game.UndoN(n) })
}

// RestoreBristolInteractor deserialises JSON into a BristolInteractor.
func RestoreBristolInteractor(data []byte, op presenter.BristolPresenter) (*BristolInteractor, error) {
	return restoreAndBuild[domain.Bristol](data, func(g *domain.Bristol) *BristolInteractor {
		return &BristolInteractor{GameBase: GameBase[interfaces.BristolGame]{Game: g}, op: op}
	})
}
