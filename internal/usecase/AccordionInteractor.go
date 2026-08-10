//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// AccordionInteractorIF アコーディオンインタラクターインタフェース
type AccordionInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Move パイルを重ねる
	Move(fromIdx, toIdx int) string
	// GiveUp ギブアップ
	// AutoComplete ヒントが示す手を尽きるまで繰り返す
	AutoComplete() string
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// AccordionInteractor アコーディオンインタラクタークラス
type AccordionInteractor struct {
	GameBase[interfaces.AccordionGame]
	ap presenter.AccordionPresenter
}

// NewAccordionInteractor コンストラクタ
func NewAccordionInteractor(a interfaces.AccordionGame, ap presenter.AccordionPresenter) *AccordionInteractor {
	mustNotNil("AccordionInteractor", map[string]any{"a": a, "ap": ap})
	return &AccordionInteractor{GameBase: GameBase[interfaces.AccordionGame]{Game: a}, ap: ap}
}

// Reset ゲーム初期化
func (ai *AccordionInteractor) Reset() string {
	return runAndPresent(ai.Game, ai.ap, ai.Game.Reset)
}

// Move パイルを重ねる
func (ai *AccordionInteractor) Move(fromIdx, toIdx int) string {
	return execAndPresent(ai.Game, ai.ap, func() error {
		return ai.Game.Move(fromIdx, toIdx)
	})
}

// AutoComplete ヒントが示す手を尽きるまで繰り返す
func (ai *AccordionInteractor) AutoComplete() string {
	return execAndPresent(ai.Game, ai.ap, ai.Game.AutoComplete)
}

// GiveUp ギブアップ
func (ai *AccordionInteractor) GiveUp() string {
	return runAndPresent(ai.Game, ai.ap, ai.Game.GiveUp)
}

// Hint ヒント取得
func (ai *AccordionInteractor) Hint() string {
	return ai.ap.HintOutput(ai.Game)
}

// ActionLog 棋譜を出力する
func (ai *AccordionInteractor) ActionLog() string {
	return ai.ap.ActionLogOutput(ai.Game)
}

// Undo アンドゥ
func (ai *AccordionInteractor) Undo() string {
	return execAndPresent(ai.Game, ai.ap, ai.Game.Undo)
}

// UndoN n回連続アンドゥ
func (ai *AccordionInteractor) UndoN(n int) string {
	return execAndPresent(ai.Game, ai.ap, func() error { return ai.Game.UndoN(n) })
}

// RestoreAccordionInteractor deserialises JSON into an AccordionInteractor.
func RestoreAccordionInteractor(data []byte, ap presenter.AccordionPresenter) (*AccordionInteractor, error) {
	return restoreAndBuild[domain.Accordion](data, func(g *domain.Accordion) *AccordionInteractor {
		return &AccordionInteractor{GameBase: GameBase[interfaces.AccordionGame]{Game: g}, ap: ap}
	})
}
