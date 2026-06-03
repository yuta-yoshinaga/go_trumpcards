//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OsmosisInteractorIF オズモシスインタラクターインタフェース
type OsmosisInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// MoveWasteToFoundation ウェイストからファンデーションに移動
	MoveWasteToFoundation(fIdx int) string
	// MoveReserveToFoundation リザーブからファンデーションに移動
	MoveReserveToFoundation(rIdx, fIdx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜出力
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// OsmosisInteractor オズモシスインタラクタークラス
type OsmosisInteractor struct {
	GameBase[interfaces.OsmosisGame]
	op presenter.OsmosisPresenter
}

// NewOsmosisInteractor コンストラクタ
func NewOsmosisInteractor(o interfaces.OsmosisGame, op presenter.OsmosisPresenter) *OsmosisInteractor {
	mustNotNil("OsmosisInteractor", map[string]any{"o": o, "op": op})
	return &OsmosisInteractor{GameBase: GameBase[interfaces.OsmosisGame]{Game: o}, op: op}
}

// Reset ゲーム初期化
func (oi *OsmosisInteractor) Reset() string {
	return runAndPresent(oi.Game, oi.op, oi.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (oi *OsmosisInteractor) Draw() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Draw)
}

// MoveWasteToFoundation ウェイストからファンデーションに移動
func (oi *OsmosisInteractor) MoveWasteToFoundation(fIdx int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.MoveWasteToFoundation(fIdx) })
}

// MoveReserveToFoundation リザーブからファンデーションに移動
func (oi *OsmosisInteractor) MoveReserveToFoundation(rIdx, fIdx int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.MoveReserveToFoundation(rIdx, fIdx) })
}

// GiveUp ギブアップ
func (oi *OsmosisInteractor) GiveUp() string {
	return runAndPresent(oi.Game, oi.op, oi.Game.GiveUp)
}

// Hint ヒント取得
func (oi *OsmosisInteractor) Hint() string {
	return oi.op.HintOutput(oi.Game)
}

// AutoComplete オートコンプリート
func (oi *OsmosisInteractor) AutoComplete() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.AutoComplete)
}

// ActionLog 棋譜出力
func (oi *OsmosisInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// Undo アンドゥ
func (oi *OsmosisInteractor) Undo() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Undo)
}

// UndoN n回連続アンドゥ
func (oi *OsmosisInteractor) UndoN(n int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.UndoN(n) })
}

// RestoreOsmosisInteractor deserialises JSON into an OsmosisInteractor.
func RestoreOsmosisInteractor(data []byte, op presenter.OsmosisPresenter) (*OsmosisInteractor, error) {
	return restoreAndBuild[domain.Osmosis](data, func(g *domain.Osmosis) *OsmosisInteractor {
		return &OsmosisInteractor{GameBase: GameBase[interfaces.OsmosisGame]{Game: g}, op: op}
	})
}
